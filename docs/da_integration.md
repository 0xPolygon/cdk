# CDK DA Integration

The purpose of this document is to explain how a 3rd Party Data Availability (DA) solution can integrate with CDK.  

## Considerations

The code outlined in this document is under development, and while we’re confident that it will be ready for production in a few weeks, it is currently under heavy development.
  
For the first iteration of integrations, on-chain verification is not expected. Although this document shows how this could be done at the contract level (doing such a thing on the ZKPs is out of the scope right now). In any case, Agglayer will assert that the data is actually available before settling ZKPs.

## Smart Contracts

The versions of the smart contracts that are being targeted for the DA integrations are found in [zkevm-contracts @ feature/banana](https://github.com/0xPolygonHermez/zkevm-contracts/tree/feature/banana). This new version of the contracts allow for multiple “consensus” implementations but there are two that are included by default:

- zkEVM to implement a rollup.
- Validium to implement a validium.
- Adding a custom solution. 

This document only considers the first approach, reusing the `PolygonValidium` consensus. That being said, the `PolygonValidium` implementation allows a custom smart contract to be used in the relevant interaction. This could be used by DAs to add custom on-chain verification logic. While verifying the DA integrity is optional, any new protocol will need to develop a custom smart contract in order to be successfully  integrated (more details bellow)

This is by far the [most relevant part of the contract for DAs](https://github.com/0xPolygonHermez/zkevm-contracts/blob/533641301223a1e413b2e8f0323354671f310922/contracts/v2/consensus/validium/PolygonValidiumEtrog.sol#L91C5-L98C36):

```javascript
    function sequenceBatchesValidium(
        ValidiumBatchData[] calldata batches,
        uint32 indexL1InfoRoot,
        uint64 maxSequenceTimestamp,
        bytes32 expectedFinalAccInputHash,
        address l2Coinbase,
        bytes calldata dataAvailabilityMessage
    ) external onlyTrustedSequencer {
```

And in particular this [piece of code](https://github.com/0xPolygonHermez/zkevm-contracts/blob/feature/banana/contracts/v2/consensus/validium/PolygonValidiumEtrog.sol#L228C13-L230):

```javascript
    // Validate that the data availability protocol accepts the dataAvailabilityMessage
    // note This is a view function, so there's not much risk even if this contract was vulnerable to reentrant attacks
    dataAvailabilityProtocol.verifyMessage(
        accumulatedNonForcedTransactionsHash,
        dataAvailabilityMessage
    );
```

This is what the data availability committee uses, but the `PolygonDataCommittee` smart contract can be modified, as long as it follows the interface. The `dataAvailabilityMessage` is an input parameter that can be used in any arbitrary way. **So DAs could use this to add custom logic on L1.**

## Setup the Node

In order to integrate a DA solution into CDK, the most fundamental part is for the node to be able to post and retrieve data from the DA backend.  

Up until now, DAs would fork the `cdk-validium-node` repo to make such an integration. But maintaining forks can be really painful, so the team is proposing this solution that will allow the different DAs to be 1st class citizens and live on the official `cdk` repo. 

These items would need to be implemented to have a successful integration:

1. Create a repository that will host the package that implements [this interface](https://github.com/0xPolygon/cdk/blob/develop/dataavailability/interfaces.go#L11-L16). You can check how is done for the [DAC case](https://github.com/0xPolygon/cdk/blob/develop/dataavailability/datacommittee/datacommittee.go) as an example.
2. Add a new entry on the [supported backend strings](https://github.com/0xPolygon/cdk/blob/develop/dataavailability/config.go)
3. [OPTIONAL] Add a config struct in the new package, and add the struct inside the main data availability config struct, this way your package will be able to receive custom configuration using the main config file of the node.
4. `go get` and instantiate your package and use it to create the main data availability instance, as done in the Polygon implementation.

> [!TIP]
> By default all E2E tests will run using the DAC. It’s possible to run the E2E test using other DA backends changing the test config file.

## Test the integration

1. Create an E2E test that uses your protocol, as it’s done for the DAC here
2. Generate a docker image with your changes in the node: make build-docker
3. Update the L1 config with values from the docker deployment (after successfully generating the docker image, a directory should be created under `docker/deploymentOutput`).
4. Update the contracts docker image with the one you tagged before (in the example hermeznetwork/geth-cdk-validium-contracts:XXXX) here
5. Modify the Makefile to be able to run your test, take the case of the DAC test as an example here

### Example flow

1. Sequencer groups N batches of arbitrary size into a sequence
2. Sequencer calls `PostSequence`
3. The DA BAckend implementation decides to split the N batches into M chunks, so they fit as good as possible to the size of the DA blobs of the protocol (or other considerations that the protocol may have)
4. The DA BAckend crafts the `dataAvailabilityMessage`, this is optional but could be used to:
    - Verify the existance of the data on the DA backend on L1 (this message will be passed down to the DA smart contract, and it could include merkle proofs, ...). Realisitcally speaking, we don't expect to be implemented on a first iteration
    - Help the data retrival process, for instance by including the block height or root of the blobs used to store the data. If many DA blobs are used to store a single sequence, one interesting trick would be to post some metadata in another blob, or the lates used blob, that points to the other used blobs. This way only the pointer to the metadata is needed to include into the `dataAvailabilityMessage` (since this message will be posted as part of the calldata, it's interesting to minimize it's size)
5. The sequencer [posts the sequence on L1](https://github.com/0xPolygonHermez/zkevm-contracts/blob/develop/contracts/v2/consensus/validium/PolygonValidiumEtrog.sol#L85), including the `dataAvailabilityMessage`. On that call, [the DA smart contract will be called](https://github.com/0xPolygonHermez/zkevm-contracts/blob/develop/contracts/v2/consensus/validium/PolygonValidiumEtrog.sol#L217). This can be used to validate that the DA protocol has been used as expected (optional)
6. After that happens, any node synchronizing the network will realise of it through an event of the smart contract, and will be able to retrieve the hashes of each batch and the `dataAvailabilityMessage`
7. And so it will be able to call `GetSequence(hashes common.Hash, dataAvailabilityMessage []byte)` to the DA Backend
8. The DA BAckend will then retrieve the data, and return it

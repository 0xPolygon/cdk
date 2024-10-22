# Non EVM Integration

The purpose of this document is to explain how a 3rd Party Execution Environment can integrate with LxLy/AggLayer using the CDK.

## Disclaimer

This is experimental, and there is no working examples of anything non EVM being integrated to the AggLayer yet. While the core principles of what needs to be done in order to achieve this are clear, the implementation details on how to get there are more than likely to change. Consider this as a reference on the amount of work that needs to be done, rather than the path to a production grade integration

## Key concepts

Anything (chain or not chain) should be able to interact with the [LxLy](https://docs.polygon.technology/zkEVM/architecture/unified-LxLy) bridge, and settle using the [AggLayer](https://docs.polygon.technology/learn/agglayer/overview/). Specially if done through the [Pesimistic Proof](https://docs.polygon.technology/learn/agglayer/pessimistic_proof/). In the future more advanced and secure options (consensus, execution, DA, ...) will be provided, but as of now, this doc assumes that the integration uses exclusively the Pesimistic Proof settlement configuration

The CDK client is in charge of handling the integration to both the LxLy bridge and the AggLayer. Consider it an SDK to bring your project to the AggLayer ecosystem. In order to do so, some code will be needed, in a adaptor/plugin fashion, for the CDK client to index and interact with the service being integrated

In some cases the integration will require to implement some code in `Go`. When this is the case, the expected procedure is for the implementaiton to exist in another repo, and imported in this repo (`cdk`) as an external dependency. In general the idea is to provide implementations that know how to interact with the "smart contracts" of the system being integrated. By doing so, the system is abstracted behind well defined APIs and interfaces, enabling the `cdk` client to re-use the same logic for all the different systems. In other words, in order to integrate a new system, some "adaptors" are needed, beyond that the existing code should take care of the rest

## Components needed for the integration

### Smart Contracts

In the case of the EVM integrations, there are two relevant smart contracts:

- [Global Exit Root](TODO)
- [Bridge](TODO)

It will be necessary for the system being integrated to implement equivalent functionality. It is not necessary that this is done in the form of a smart contract (altho this doc will refere to this as a smart contract for simplicity) or that it's splitted in two different entities, but the functionality should be respected. tl;dr of the functionality:

- Bridge assets and messages to other networks
- Claim assets / messages (incomming bridges)
- Export Local Exit Roots: hash that will be needed for other networks to claim the assets
- Import Global Exit Roots: hash uppon the bridges are claimed

### AggOracle

This component is in charge of importing the Global Exit Roots into the smart contract(s). This should be implemented in the form of a `Go` package. Take the [EVM implementation](../aggoracle/chaingersender/evm.go) as reference. This implementation should follow the interface `ChainSender` defined [here](../aggoracle/oracle.go)

### BridgeSync

This component is responsible for synchronizing the bridge information, in particular for this context, the `bridges` and `claims` originated on the L2 attached to the cdk client (aka "your service"). In other words, the responsability of this component will be to monitor what's happening on the bridge smart contract. This is necessary in order to collect the data needed to interact with the AggLayer, but also to feed the bridge service, enabling users to claim bridges originated on this network to the destination network

> WARNING: The following interfaces are likely to change

**self note:** it's not very nice to have the "ProcessorInterface" as integration point since it hides the actual events in another interface, making things confusing

In order to process the activity originated on the "non-EVM contracts" a `downloader` and `driver` will be needed. The current implementation needs some adjustments in order to accomodate custom implementations, but in a nutshell, what needs to be done is to interact with the [processor](../bridgesync/processor.go), in particular the `ProcessorInterface` found [here](../sync/driver.go) should be implemented. Note that the `Events` in `Block` is just a interface, in this case, this interface should be able to be parsed as a `Event` struct as defined on the [processor](../bridgesync/processor.go).

### Claim sponsor

This component is responsible for performing claims on behalf of the user. This is specially necessary on systems that have the concept of "gas" (transactions are paid). This is the case because otherwise those gas based systems could face a chicken/egg situation: how can a user pay for the claim transaction, if it needs to have previously done a claim transaction in order to have funds to pay for it.

It's important to note that this component is completley optional, and in fact may be pointless on some systems. In fact the [bridge RPC](../rpc/bridge.go) where the claim sponsor is consumed, has a config parameter to enable/disable it

In order to implement a claim sponsor that is capable to perform claim transactions on the bridge smart contract, the `ClaimSender` interface needs to be implemented. This interface is defined [here](../claimsponsor/claimsponsor.go)

### Last GER sync

> WARNING: The following interfaces are likely to change

This component is in charge of undertanding what Global Exit Roots have been imported on the system. This is important for the bridge service to understand when incomming bridges are ready to be claimed

In terms of what needs to be done, it's pretty similar to the `Bridge Sync` process, the [ProcessorInterface](../sync/driver.go) needs to be implemented, this time the events should be of the type `Event` as defined [here](../lastgersync/processor.go).

## Considerations

### Bridge

Once the previously mentioned components are fully implemented, the network being integrated should be fully plugged into the LxLy bridge, however there are few things worth noticing:

- While bridges going from the network being integrated to other networks are expected to work with the existing tooling and UIs, this is not the case for the opposite flow. This is the case due to the fact that current UI implementations only know how to perform claims on EVM networks. That being said, the auto claim feature overcomes this situation as it standerizes a way for claiming across different systems. However, any other way of "claiming" is left open to be implemented by the sytem being integrated
- Something similar happens to "bridging" (the interaction of sending assets/messages to another network). This is specific to the system being integrated, and therefore is left open for the system being integrated to add mechanisms to interact with the "bridge smart contract" in order to trigger those actions
- We're transitioning to use an "in CDK" bridge service implementation (spec drafted [here](https://hackmd.io/0vA-XU2BRHmH3Ab0j4ouZw)) as opposite to the current separated service implemented [here](https://github.com/0xPolygonHermez/zkevm-bridge-service). This means that there is not yet an stable API and other tooling such as SDKs and UIs are missing

### AggLayer

The AggLayer integration will fully work once the previously mentioned components are implemented, but it's worth noticing that this integration will only include the Pesimistic Proof. In future iterations, the security of the systems being integrated will be increased by adding other proofs and features such as: execution proof, consensus proof, DA proof, forced transactions, ... All those upcomming improvements will be optional, while the Pesimistic Proof will remain mandatory regardless of the use of other proofs
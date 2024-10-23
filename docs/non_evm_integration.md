# Integrating non-EVM systems

This guide explains how to connect a third-party execution environment to the AggLayer using the CDK.

## Important note

The following information is experimental, and there aren't any working examples of non-EVM integrations with the AggLayer yet. While we know what needs to be done conceptually, the implementation details are likely to evolve. Think of this as a rough overview of the effort involved, rather than a step-by-step guide towards a production deployment.

## Key Concepts

Any system (chain or not chain) should be able to interact with the [unified LxLy bridge](https://docs.polygon.technology/zkEVM/architecture/unified-LxLy) and settle using the [AggLayer](https://docs.polygon.technology/learn/agglayer/overview/); especially when using the [Pessimistic Proof](https://docs.polygon.technology/learn/agglayer/pessimistic_proof/) option. Support for additional proofs, such as consensus, execution, or data availability are planned for the future. But, for now, this guide is based solely on using the Pessimistic Proof for settlement.

The CDK client handles the integration with both the unified LxLy bridge and AggLayer. Think of it as an SDK to bring your project into the AggLayer ecosystem. You'll need to write some custom code in an adapter/plugin style so that the CDK client can connect with your service.

In some cases, you might need to write code in `Go`. When that happens, the code should be in a separate repo and imported into the CDK as a dependency. The goal is to provide implementations that can interact with the *smart contracts* of the system being integrated, allowing the CDK client to reuse the same logic across different systems. Basically, you’ll need to create some *adapters* for the new system, while the existing code handles the rest.

## Components for integration

### Smart contracts

For EVM-based integrations, there are two relevant smart contracts:

- [Global exit root](https://github.com/0xPolygonHermez/zkevm-contracts/blob/feature/sovereign-bridge/contracts/v2/sovereignChains/GlobalExitRootManagerL2SovereignChain.sol)
- [Bridge](https://github.com/0xPolygonHermez/zkevm-contracts/blob/feature/sovereign-bridge/contracts/v2/sovereignChains/BridgeL2SovereignChain.sol)

The integrated system needs to implement similar functionality. It doesn't have to be a smart contract per se, and it doesn't need to be split into two parts, but it should perform the functions that we list here:

- Bridge assets and messages to other networks.
- Handle incoming asset/message claims.
- Export local exit roots (a hash needed for other networks to claim assets).
- Import global exit roots (a hash needed for processing bridge claims).

### AggOracle

This component imports global exit roots into the smart contract(s). It should be implemented as a `Go` package, using the [EVM example](../aggoracle/chaingersender/evm.go) as a reference. It should implement the `ChainSender` interface defined [here](../aggoracle/oracle.go).

### BridgeSync

BridgeSync synchronizes information about bridges and claims originating from the L2 service attached to the CDK client. In other words, it monitors what's happening with the bridge smart contract, collects the necessary data for interacting with the AggLayer, and feeds the bridge service to enable claims on destination networks.

> **Heads up:** These interfaces may change.

To process events from non-EVM systems, you'll need a `downloader` and `driver`. The current setup needs some tweaks to support custom implementations. In short, you need to work with the [`Processor`](../bridgesync/processor.go), particularly the `ProcessorInterface` found [here](../sync/driver.go). The `Events` in `Block` are just interfaces, which should be parsed as `Event` structs defined in the [`Processor`](../bridgesync/processor.go).

### Claim sponsor

This component performs claims on behalf of users, which is crucial for systems with "gas" fees (transaction costs). Without it, gas-based systems could face a chicken/egg situation: How can users pay for a claim if they need a previous claim to get the funds to pay for it?

The claim sponsor is optional and may not be needed in some setups. The [bridge RPC](../rpc/bridge.go) includes a config parameter to enable or disable it. To implement a claim sponsor that can perform claim transactions on the bridge smart contract, you'll need to implement the `ClaimSender` interface, defined [here](../claimsponsor/claimsponsor.go).

### Last GER sync

> **Warning:** These interfaces may also change.

This component tracks which global exit roots have been imported. It helps the bridge service know when incoming bridges are ready to be claimed. The work needed is similar to that for the bridge sync: Implement the [`ProcessorInterface`](../sync/driver.go), with events of type `Event` defined [here](../lastgersync/processor.go).

## Additional considerations

### Bridge

Once all components are implemented, the network should be connected to the unified LxLy bridge. However, keep in mind:

- Outgoing bridges should work with current tools and UIs, but incoming bridges may not. When using the claim sponsor, things should just work. However, the claim sponsor is optional... The point being that the existing UIs are built to send EVM transactions to make the claim in the absence of claim sponsor. So any claim interaction beyond the auto-claim functionality will need UIs and tooling that are out of the sope of the CDK.
- Bridging assets/messages to another network is specific to the integrated system. You'll need to create mechanisms to interact with the *bridge smart contract* of your service for these actions.
- We’re moving towards an *in-CDK* bridge service (spec [here](https://hackmd.io/0vA-XU2BRHmH3Ab0j4ouZw)), replacing the current separate service ([here](https://github.com/0xPolygonHermez/zkevm-bridge-service)). There's no stable API yet, and SDKs/UIs are still in development.

### AggLayer

AggLayer integration will work once the components are ready, but initially, it will only support Pessimistic Proof. Later updates will add more security features like execution proofs, consensus proofs, data availability, and forced transactions. These will be optional, while Pessimistic Proof will remain mandatory.
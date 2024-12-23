# AggOracle Component - Developer Documentation

## Overview

The **AggOracle** component ensures the **Global Exit Root (GER)** is propagated from L1 to the L2 sovereign chain smart contract. This is critical for enabling asset and message bridging between chains.

The GER is picked up from the smart contract by **LastGERSyncer** for local storage.

### Key Components:

- **ChainSender**: Interface for submitting GERs to the smart contract.
- **EVMChainGERSender**: An implementation of `ChainSender`.

---

## Workflow

### What is Global Exit Root (GER)?

The **Global Exit Root** consolidates:

- **Mainnet Exit Root (MER)**: Updated during bridge transactions from L1.
- **Rollup Exit Root (RER)**: Updated when verified rollup batches are submitted via ZKP.

        GER = hash(MER, RER)

### Process

1. **Fetch Finalized GER**:
    - AggOracle retrieves the latest GER finalized on L1.
2. **Check GER Injection**:
    - Confirms whether the GER is already stored in the smart contract.
3. **Inject GER**:
    - If missing, AggOracle submits the GER via the `insertGlobalExitRoot` function.
4. **Sync Locally**:
    - LastGERSyncer fetches and stores the GER locally for downstream use.

The sequence diagram below depicts the interaction in the AggOracle.

```mermaid
sequenceDiagram
    participant AggOracle
    participant ChainSender
    participant L1InfoTreer
    participant L1Client

    AggOracle->>AggOracle: start
    loop trigger on preconfigured frequency
        AggOracle->>AggOracle: process latest GER
        AggOracle->>L1InfoTreer: get last finalized GER
        alt targetBlockNum == 0
            AggOracle->>L1Client: get (latest) header by number
            L1Client-->>AggOracle: the latest header
            AggOracle->>L1InfoTreer: get latest L1 info tree until provided header
            L1InfoTreer-->>AggOracle: global exit root (from L1 info tree)
        else
            AggOracle->>L1InfoTreer: get latest L1 info tree until provided header
            L1InfoTreer-->>AggOracle: global exit root (from L1 info tree)
        end
        AggOracle->>ChainSender: check is GER injected
        ChainSender-->>AggOracle: isGERInjected result
        alt is GER injected
            AggOracle->>AggOracle: log GER already injected
        else
            AggOracle->>ChainSender: inject GER
            ChainSender-->>AggOracle: GER injection result
        end
    end
    AggOracle->>AggOracle: handle GER processing error
```

---

## Key Components

### 1. AggOracle

The `AggOracle` fetches the finalized GER and ensures its injection into the L2 smart contract.

### Functions:

- **`Start`**: Periodically processes GER updates using a ticker.
- **`processLatestGER`**: Checks if the GER exists and injects it if necessary.
- **`getLastFinalizedGER`**: Retrieves the latest finalized GER based on block finality.

---

### 2. ChainSender Interface

Defines the interface for submitting GERs.

```
IsGERInjected(ger common.Hash) (bool, error)
InjectGER(ctx context.Context, ger common.Hash) error
```

---

### 3. EVMChainGERSender

Implements `ChainSender` using Ethereum clients and transaction management.

### Functions:

- **`IsGERInjected`**: Verifies GER presence in the smart contract.
- **`InjectGER`**: Submits the GER using the `insertGlobalExitRoot` method and monitors transaction status.

---

## Smart Contract Integration

- **Contract**: `GlobalExitRootManagerL2SovereignChain.sol`
- **Function**: `insertGlobalExitRoot`
    - [Source Code](https://github.com/0xPolygonHermez/zkevm-contracts/blob/feature/audit-remediations/contracts/v2/sovereignChains/GlobalExitRootManagerL2SovereignChain.sol#L89-L103)
- **Bindings**: Available in [cdk-contracts-tooling](https://github.com/0xPolygon/cdk-contracts-tooling/tree/main/contracts/l2-sovereign-chain).

---

## Summary

The **AggOracle** component automates the propagation of GERs from L1 to L2, enabling bridging across networks.

Refer to the EVM implementation in [evm.go](https://github.com/0xPolygon/cdk/blob/main/aggoracle/chaingersender/evm.go) for guidance on building new chain senders.

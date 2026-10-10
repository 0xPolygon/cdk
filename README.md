> [!WARNING]
> **This repository is deprecated and no longer maintained.**
>
> The legacy CDK node stack (CDK node, validium node, sequence sender, data availability committee) has been retired. Polygon CDK chains now run on
> [cdk-op-reth](https://github.com/0xPolygon/cdk-op-reth) and settle to the
> [Agglayer](https://github.com/agglayer/agglayer) with pessimistic proofs and
> full execution proofs. This code is kept for reference only. It will not
> receive bug fixes, security patches, or releases.
>
> **Use instead:**
> - Connecting a chain to the Agglayer: [agglayer/aggkit](https://github.com/agglayer/aggkit)
> - Execution client / node: [0xPolygon/cdk-op-reth](https://github.com/0xPolygon/cdk-op-reth)
> - Docs: [docs.polygon.technology](https://docs.polygon.technology)
>
> Security issues in live Polygon systems: see [SECURITY.md](./SECURITY.md).

<div id="top"></div>
<!-- PROJECT LOGO -->
<br />
<div align="center">

<img src="./.github/assets/cdk-logo.svg#gh-light-mode-only" alt="Logo" width="100">
<img src="./.github/assets/cdk-logo.svg#gh-dark-mode-only" alt="Logo" width="100">

## Polygon CDK

**Polygon CDK** (Chain Development Kit) is a modular framework that developers can use to build and deploy Zero Knowledge Proofs enabled Rollups and Validiums.

The CDK allow to build Rollups that are ZK powered, verifying the execution using the zkEVM prover from Polygon, they can be completely personalizable because its modullar architecture.

<!-- PROJECT SHIELDS -->
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=0xPolygon_cdk&metric=alert_status&token=aa6d76993fc213c4153bf65e0d62e4d08207ea7e)](https://sonarcloud.io/summary/new_code?id=0xPolygon_cdk)

</div>
</div>

<br />

## Getting Started

## Pre-requisites

Setup Kurtosis following this instructions https://github.com/0xPolygon/kurtosis-cdk?tab=readme-ov-file#getting-started

### Local Testing

- You can run locally against kurtosis-cdk environment using: [docs/local_debug.md](docs/local_debug.md)

### Build locally

You can locally build a production release of CDK CLI + cdk-node with:

```
make build
```

### Run locally

You can build and run a debug release locally using:

```
cargo run
```

It will build and run both binaries.
### Running with Kurtosis

1. Run your kurtosis environment
2. build `cdk-erigon` and make it available in your system's PATH
3. Run `scripts/local_config`
4. cargo run -- --config ./tmp/cdk/local_config/test.kurtosis.toml --chain ./tmp/cdk/local_config/genesis.json erigon

## Contributing

Contributions are very welcomed, the guidelines are currently not available (WIP)

## Support

Feel free to [open an issue](https://github.com/0xPolygon/cdk/issues/new) if you have any feature request or bug report.<br />


## License

Polygon Chain Development Kit
Copyright (c) 2024 PT Services DMCC

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

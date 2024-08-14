# Setup environment to local debug on VSCode

## Requirements

* Working and running [kurtosis-cdk](https://github.com/0xPolygon/kurtosis-cdk) environment setup.
* In `test/scripts/env.sh` setup `KURTOSIS_FOLDER` pointing to your setup.

> [!TIP]
> Use your WIP branch in Kurtosis as needed

## Create configuration for this kurtosis environment

```
scripts/local_config
```

## Stop cdk-node started by Kurtosis

```bash
kurtosis service stop cdk-v1 cdk-node-001
```

## Add to vscode launch.json

After execution `scripts/local_config` it suggest an entry for `launch.json` configurations

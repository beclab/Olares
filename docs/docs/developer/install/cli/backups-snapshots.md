---
outline: [2, 3]
---
# `snapshots`
:::warning
The `olares-cli backups download` subcommand must be run first to install Restic. Otherwise, this command will return an error.
:::
## Synopsis
The `snapshots` subcommand lists all available snapshots for a specific backup repository. It supports multiple storage backends and provides the necessary options to authenticate and query snapshots.

```bash
olares-cli backups snapshots <backend> --repo-name <name> [options]
```
## Common options
These options apply to all backends:

| Option      | Shorthand | Usage                                                 | Required | Default |
|---------------|-----------|-------------------------------------------------------|-------------------------|---------|
| `--help`      | `-h`      | Displays help information.                            | No                   | N/A     |
| `--repo-name` |           | Specifies the name of the backup repository to query. | No                   | N/A     |

## Backend-specific options

### Options for `cos`

| Option              | Shorthand | Usage                                                                                            | Required | Default |
|---------------------|-----------|--------------------------------------------------------------------------------------------------|-------------------------|---------|
| `--access-key`      |           | Specifies the Access Key for Tencent COS.                                                        | No                   | N/A     |
| `--endpoint`        |           | Specifies the Tencent COS endpoint, e.g., `https://cos.{region}.myqcloud.com/{bucket}/{prefix}`. | No                   | N/A     |
| `--secret-access-key` |           | Specifies the Secret Access Key for Tencent COS.                                                 | No                   | N/A     |

### Options for `fs`

| Option     | Shorthand | Usage                                                          | Required | Default |
|------------|-----------|----------------------------------------------------------------|-------------------------|---------|
| `--endpoint` |           | Specifies the local directory where the backup will be stored. | No                   | N/A     |

### Options for `s3`

| Option              | Shorthand | Usage                                                                                       | Required | Default |
|---------------------|-----------|---------------------------------------------------------------------------------------------|-------------------------|---------|
| `--access-key`      |           | Specifies the Access Key for Amazon S3.                                                     | No                   | N/A     |
| `--endpoint`        |           | Specifies the Amazon S3 endpoint, e.g., `https://{bucket}.{region}.amazonaws.com/{prefix}`. | No                   | N/A     |
| `--secret-access-key` |           | Specifies the Secret Access Key for Amazon S3.                                              | No                   | N/A     |

### Options for `space`

| Option                        | Shorthand | Usage                                                                                                                                    | Required | Default |
|-------------------------------|-----------|------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|---------|
| `--access-token` <sup>1</sup> |           | Specifies the access token for Olares Space.                                                                                             | No                   | N/A     |
| `--cloud-api-mirror`          |           | Specifies the cloud API mirror.                                                                                                          | No                   | N/A     |
| `--cloud-name`                |           | Specifies the cloud name of the Olares Space instance. <br/> The cloud name can be retrieved using the [`region`](./backups-region.md) subcommand. | No                   | N/A     |
| `--cluster-id` <sup>2</sup>   |           | Specifies the cluster ID where the backup will be stored.                                                                                | No                   | N/A     |
| `--olares-did` <sup>1</sup>   |           | Specifies the Olares DID.                                                                                                                | No                   | N/A     |
| `--region-id`                 |           | Specifies the region ID of the Olares Space instance. <br/> The region ID can be retrieved using the [`region`](./backups-region.md) subcommand.   | No                   | N/A     |

1. To retrieve the access token and Olares DID, inspect the payload of the network requests made by the Olares Space web interface after logging in. The `token` field corresponds to the access token, and the `userid` field corresponds to the Olares DID.

2. To retrieve the cluster ID, use the following command:
  ```bash
  kubectl get terminus -o jsonpath='{.items[*].metadata.labels.bytetrade\.io/cluster-id}'
  ```

## Example
```bash
# List snapshots for Tencent COS
olares-cli backups snapshots cos --repo-name my_repo \
  --access-key YOUR_KEY \
  --secret-access-key YOUR_SECRET \
  --endpoint https://cos.region.myqcloud.com/bucket/prefix
  
# List snapshots for local filesystem
olares-cli backups snapshots fs --repo-name my_repo --endpoint /backup_repo
```
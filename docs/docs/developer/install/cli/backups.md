# `backups`
The `backups` command provides a set of tools to manage data backups, restores, and snapshots. It supports multiple storage backends, including Tencent COS, Amazon S3, the local filesystem, and Olares Space.

## Subcommands

| Subcommand  | Description                                                                                              |
|-------------|----------------------------------------------------------------------------------------------------------|
| `download`  | Downloads the Restic dependency tool.                                                                    |
| `region`    | Retrieves the cloud name and region ID. Specifically used only when the storage backend is Olares Space. |
| `backup`    | Backups data to a specified storage backend.                                                             |
| `restore`   | Restores data from a specified storage backend.                                                          |
| `snapshots` | Manages and views backup snapshots.                                                                      |

## Available backends

The `<backend>` parameter specifies the storage backend for the `backup` and `restore` commands. Olares CLI supports the following backends:

| Backend | Description                                                                           |
|---------|---------------------------------------------------------------------------------------|
| `cos`   | Tencent Cloud Object Storage (COS). Requires an access key, secret key, and endpoint. |
| `s3`    | Amazon Simple Storage Service (S3). Requires an access key, secret key, and endpoint. |
| `fs`    | Local filesystem. No credentials required.                                            |
| `space` | Olares Space. Requires an access token.                                               |
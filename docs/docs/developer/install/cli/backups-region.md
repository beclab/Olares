# `region`

## Synopsis
The `region` subcommand is used to retrieve the cloud name and region ID. It is specifically used when the storage backend is Olares Space.
```bash
olares-cli backups region space [options]
```

## Options

| Option                        | Shorthand | Usage                                        | Required | Default |
|-------------------------------|-----------|----------------------------------------------|-------------------------|---------|
| `--access-token` <sup>1</sup> |           | Specifies the access token for Olares Space. | No                   | N/A     |
| `--cloud-api-mirror`          |           | Specifies the cloud API mirror.              | No                   | N/A     |
| `--help`                      | `-h`      | Displays help information.                   | No                   | N/A     |
| `--olares-did` <sup>1</sup>   |           | Specifies the Olares DID.                    | No                   | N/A     |

1. To retrieve the access token and Olares DID, inspect the payload of the network requests made by the Olares Space web interface after logging in. The `token` field corresponds to the access token, and the `userid` field corresponds to the Olares DID.

## Example
```bash
# Query cloud name and region ID
olares-cli backups region space \
  --access-token YOUR_ACCESS_TOKEN \
  --cloud-api-mirror https://api-mirror.example.com \
  --olares-did did:xyz123
```
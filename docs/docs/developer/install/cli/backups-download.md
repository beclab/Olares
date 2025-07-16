# `download`

## Synopsis
The `download` subcommand downloads the Restic dependency tool. Restic is required for performing backup and restore operations, as well as managing snapshots.

```bash
olares-cli backups download [options]
```
## Options

| Option             | Shorthand | Usage                                                  | Required | Default            |
|--------------------|-----------|--------------------------------------------------------|-------------------------|--------------------|
| `--download-cdn-url`|           | Specifies the CDN URL for downloading the Restic tool. | No                   | System default URL |
| `--help`           | `-h`      | Displays help information.                             | No                   | N/A                |
## Example
```bash d
# Download Restic using a custom CDN URL
olares-cli backups download --download-cdn-url https://custom-cdn.example.com/restic
```
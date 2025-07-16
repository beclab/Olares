# `download`

## Synopsis
The `download` command downloads the necessary packages and components required to install Olares on your local machine. It supports downloading components, checking the status of installation packages, and fetching the manifest file.

```bash
olares-cli download <subcommand> [option]
```

## Subcommands

| Name        | Shorthand | Usage                                                 | Example                                |
|-------------|-----------|-------------------------------------------------------|----------------------------------------|
| `check`     |           | Checks the status of the Olares installation package. | `olares-cli download check`     |
| `component` | `c`       | Downloads Olares components.                          | `olares-cli download component` |
| `wizard`    | `w`       | Downloads the manifest file.                          | `olares-cli download wizard`    |

## Options

| Option             | Shorthand | Usage                                                                                                                                                                                                                                                     | Required | Default                        |
|--------------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|--------------------------------|
| `--base-dir`       | `-b`      | Sets the base directory for Olares.                                                                                                                                                                                                                       | No                   | `$HOME/.olares`                |
| `--download-cdn-url`|           | Sets the CDN accelerated download URL in the format `https://example.cdn.com`.                                                                                                                                                                            | No                   | System default URL             |
| `--help`           | `-h`      | Displays help information.                                                                                                                                                                                                                                | No                   | N/A                            |
| `--kube`           |           | Specifies the Kubernetes type. <br>Supported types are `k3s` and `k8s`.                                                                                                                                                                                   | No                   | `k3s`                          |
| `--version`        | `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`). Defaults to the current version.<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. | No                   | Current version    |

## Examples
```bash
# Specifies the base directory where all downloaded components will be stored.
olares-cli download component -b /custom/path

# Specifies a CDN-accelerated URL for downloading Olares components.
olares-cli download component --download-cdn-url https://my.cdn.com

# Specifies the Kubernetes type for the installation.
olares-cli download component --kube k8s

# Sets the path to the package manifest file.
olares-cli download component --manifest /custom/path/manifest.json

# Specifies the version of Olares packages and components to download.
olares-cli download component -v 1.11.0
```



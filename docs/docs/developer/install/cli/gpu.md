# `gpu`

## Synopsis

The `gpu` command manages GPU-related operations, including installing, enabling, disabling, and uninstalling GPU drivers and related components, as well as checking the GPU status.

```bash
olares-cli gpu <subcommand> [options]
```

:::info
- By default, the Olares installation script detects your GPU hardware and CUDA drivers, then configures and enables the GPU components and services automatically.
- Currently, only NVIDIA GPUs are supported.
:::

## Subcommands

| Subcommand  | Description                                                                                                                                |
|-------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| `install`   | Installs GPU drivers and dependencies.  |
| `enable`    | Enables GPU functionality to support GPU-based applications.                                                                               |
| `disable`   | Disables GPU functionality, stopping support for GPU-based applications.                                                                   |
| `uninstall` | Uninstalls GPU drivers and related components.                                                                                             |
| `status`    | Displays the installed GPU driver version, CUDA version, and the status of GPU-related services.                                           |


## Options

| Option     | Shorthand | Usage                                                                                                                                                                                                                                                                              | Required | Default                        |
|------------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|--------------------------------|
| `--base-dir`| `-b`      | Specifies the base installation directory for the GPU components.                                                                                                                                                                                                                  | No                   | `$HOME/.olares`                |
| `--version`| `-v`      | Specifies the Olares version for GPU drivers and components. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. | No                   | Current version    |
| `--help`   | `-h`      | Displays help information.                                                                                                                                                                                                                                                         | No                   | N/A                            |

## Example

```bash
# Install GPU drivers and dependencies to a specific directory
olares-cli gpu install --base-dir /home/olares/.olares --version 1.11.1-rc.4

# Enable GPU functionality
olares-cli gpu enable

# View the status of GPU drivers and services
olares-cli gpu status

# Disable GPU functionality
olares-cli gpu disable

# Uninstall GPU drivers and components
olares-cli gpu uninstall
```



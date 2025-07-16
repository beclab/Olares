# `olares change-ip`
:::warning When manually updating IP address is required
When Olares is deployed _inside_ a virtualized environment, such as macOS (via Minikube) or Windows (via WSL), a change in the host system's IP address (e.g., due to switching Wi-Fi networks) may cause Olares to become inaccessible. This happens because the NAT gateway and DNS configuration no longer match the new IP. In such cases, you need to manually update the IP address to ensure that Olares can route traffic correctly.
  :::

## Synopsis
Change the local IP address for all Olares components.

```bash
olares-cli olares change-ip [option]
```

## Options

| Name             | Shorthand | Usage                                                                                                                                                                                                                                                     |
|------------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--base-dir`     | `-b`      | Sets the base directory for Olares.<br> Defaults to `$HOME/.olares`.                                                                                                                                                                                  |
| `--distribution` | `-d`      | Sets the WSL distribution name. Only applicable on Windows. <br> Defaults to `Ubuntu`.                                                                                                                                                                    |
| `--help`         | `-h`      | Displays help information.                                                                                                                                                                                                                                |
| `--profile`      | `-p`      | Sets the Minikube profile name. Only applicable on macOS. <br> Defaults to `olares-0`.                                                                                                                                                                    |
| `--version`      | `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. |

## Examples
- For macOS:
```bash
# Specify the Minikube profile name and change the IP.
olares-cli olares change-ip --profile olares-0
```
- For Windows WSL:
```bash
# Specify the Linux distribution in WSL and change the IP.
olares-cli olares change-ip --distribution Ubuntu
```
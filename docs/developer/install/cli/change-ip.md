# `change-ip`
:::warning When manually updating IP address is required
When Olares is deployed _inside_ a virtualized environment, such as macOS (via Minikube) or Windows (via WSL), a change in the host system's IP address (e.g., due to switching Wi-Fi networks) may cause Olares to become inaccessible. This happens because the NAT gateway and DNS configuration no longer match the new IP. In such cases, you need to manually update the IP address to ensure that Olares can route traffic correctly.
  :::

## Synopsis
Change the local IP address for all Olares components.

```bash
olares-cli change-ip [option]
```

## Options

| Option              | Shorthand | Usage                                                                                                                                                                                                                                                     | Required | Default                        |
|---------------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|--------------------------------|
| `--base-dir`        | `-b`      | Sets the base directory for Olares.                                                                                                                                                                                                                       | No                   | `$HOME/.olares`                |
| `--distribution`    | `-d`      | Sets the WSL distribution name. Only applicable on Windows.                                                                                                                                                                                             | No                   | `Ubuntu`                       |
| `--help`            | `-h`      | Displays help information.                                                                                                                                                                                                                                | No                   | N/A                            |
| `--new-master-host` |           | Specifies the new IP address of the master node on the worker node.<br> Only applicable for a multi-node Olares cluster.                                                                                                                                  | No                   | N/A                            |
| `--profile`         | `-p`      | Sets the Minikube profile name. Only applicable on macOS.                                                                                                                                                                                                 | No                   | `olares-0`                     |
| `--version`         | `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. | No                   | Current version    |

## Examples
- For macOS:
```bash
# Specify the Minikube profile name and change the IP.
olares-cli change-ip --profile olares-0
```
- For Windows WSL:
```bash
# Specify the Linux distribution in WSL and change the IP.
olares-cli change-ip --distribution Ubuntu
```
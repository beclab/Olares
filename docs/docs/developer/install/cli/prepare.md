# `prepare`

## Synopsis
The `prepare` command sets up the environment required for Olares to function. This includes installing dependencies, importing container images, and starting the Olares daemon (`olaresd`).
```bash
olares-cli prepare [option]
```

## Options
| Option             | Shorthand | Usage                                                                                                                                                                                                                                                     | Required | Default                           |
|--------------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|-----------------------------------|
| `--base-dir`       | `-b`      | Sets the base directory for Olares.                                                                                                                                                                                                                       | No                   | `$HOME/.olares`                   |
| `--help`           | `-h`      | Displays help information.                                                                                                                                                                                                                                | No                   | N/A                               |
| `--kube`           |           | Specifies the Kubernetes type. <br>Supported types are `k3s` and `k8s`.                                                                                                                                                                                   | No                   | `k3s`                             |
| `--profile`        | `-p`      | Sets the Minikube profile name. Only applicable on macOS.                                                                                                                                                                                                 | No                   | `olares-0`                        |
| `--registry-mirrors`| `-r`     | Specifies Docker container registry mirrors. <br> Multiple mirrors should be separated by commas.                                                                                                                                                         | No                   | N/A                               |
| `--with-juicefs`   |           | Configures JuiceFS as the root filesystem (rootfs) for Olares workloads instead of the local disk.                                                                                                                                                        | No                   | N/A                               |
| `--version`        | `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. | No                   | Current version |


## Example
```bash
# Uses JuiceFS as the root filesystem
olares-cli prepare --with-juicefs=true
```
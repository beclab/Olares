# `olares prepare`

## Synopsis
The `olares prepare` command sets up the environment required for Olares to function. This includes installing dependencies, importing container images, and starting the Olares daemon (`olaresd`).
```bash
olares-cli olares prepare [option]
```

## Options

| Name                 | Shorthand | Usage                                                                                                                                                                                                                                                     |
|----------------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--base-dir`         | `-b`      | Sets the base directory for Olares.<br> Defaults to `$HOME/.olares`.                                                                                                                                                                                      |
| `--help`             | `-h`      | Displays help information.                                                                                                                                                                                                                                |
| `--kube`             |           | Specifies the Kubernetes type. <br>Supported types are `k3s` (default) and `k8s`.                                                                                                                                                                         |
| `--profile`          | `-p`      | Sets the Minikube profile name. Only applicable on macOS. <br> Defaults to `olares-0`.                                                                                                                                                                    |
| `--registry-mirrors` | `-r`      | Specifies Docker container registry mirrors. <br> Multiple mirrors should be separated by commas.                                                                                                                                                         |
| `--with-juicefs`     |           | Configures JuiceFS as the root filesystem (rootfs) for Olares workloads instead of the local disk.                                                                                                                                                        |
| `--version`          | `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. |

## Example
```bash
# Uses JuiceFS as the root filesystem
olares-cli olares prepare --with-juicefs=true
```
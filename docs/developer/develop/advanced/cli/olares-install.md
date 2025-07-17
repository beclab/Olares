# `olares install`

## Synopsis
The `olares install` command installs Olares on your system. It supports various options to customize the installation process, such as specifying directories, Kubernetes types, or Minikube profiles.

```bash
olares-cli olares install [option]
```

## Options

| Name         | Shorthand | Usage                                                                                                                                                                                                                                                 |
|--------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--base-dir` | `-b`      | Sets the base directory for Olares.<br> Defaults to `$HOME/.olares`.                                                                                                                                                                                  |
| `--help`     | `-h`      | Displays help information.                                                                                                                                                                                                                            |
| `--kube`     |           | Specifies the Kubernetes type. <br>Supported types are `k3s` (default) and `k8s`.                                                                                                                                                                     |
| `--profile`  | `-p`      | Sets the Minikube profile name. Only applicable on macOS. <br> Defaults to `olares-0`.                                                                                                                                                                |
| `--version`  | `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. |


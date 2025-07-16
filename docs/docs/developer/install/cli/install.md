# `install`

## Synopsis
The `install` command installs Olares on your system. It supports various options to customize the installation process, such as specifying directories, Kubernetes types, or Minikube profiles. 

```bash
olares-cli install [option]
```

## Options

| Option     | Shorthand | Usage                                                                                                                                                                                                                                                     | Required | Default                                  |
|------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|------------------------------------------|
| `--base-dir`| `-b`      | Sets the base directory for Olares.                                                                                                                                                                                                                       | No                   | `$HOME/.olares`                          |
| `--help`   | `-h`      | Displays help information.                                                                                                                                                                                                                                | No                   | N/A                                      |
| `--kube`   |           | Specifies the Kubernetes type. <br>Supported types are `k3s` and `k8s`.                                                                                                                                                                                   | No                   | `k3s`                                    |
| `--profile`| `-p`      | Sets the Minikube profile name. Only applicable on macOS.                                                                                                                                                                                                 | No                   | `olares-0`                               |
| `--version`| `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. | No                   | Version installed in the `prepare` phase |
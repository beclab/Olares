# `uninstall`

## Synopsis
Uninstall Olares from your machine.

```bash
olares-cli uninstall [option]
```

## Options

| Option      | Shorthand | Usage                                                                                                                                                                                            | Required | Default                        |
|-------------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|--------------------------------|
| `--all`     |           | Uninstalls Olares completely, including dependencies installed during the "prepare" phase.                                                                                                       | No                   | Not applicable                 |
| `--base-dir`| `-b`      | Sets the base directory for Olares.                                                                                                                                                              | No                   | `$HOME/.olares`                |
| `--help`    | `-h`      | Displays help information.                                                                                                                                                                       | No                   | Not applicable                 |
| `--phase`   |           | Uninstalls Olares from a specific phase and revert to the previous one. <br> For example, `--phase install` removes tasks performed during the "install" phase, reverting the system to the "prepare" stage. | No                   | `install`                      |
| `--quiet`   |           | Enables quiet mode (suppress output).                                                                                                                                                            | No                   | `false`                        |
| `--version` | `-v`      | Specifies the Olares version to uninstall. <br>Use `olares-cli info` to check the downloaded version first.                                                                                         | No                   | Current version    |
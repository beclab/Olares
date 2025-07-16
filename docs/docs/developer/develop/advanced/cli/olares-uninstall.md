# `olares uninstall`

## Synopsis
Uninstall Olares from your machine.

```bash
olares-cli olares uninstall [option]
```

## Options

| Name         | Shorthand | Usage                                                                                                                                                                                                                                   |
|--------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--all`      |           | Uninstalls Olares completely, including dependencies installed during the "prepare" phase.                                                                                                                                              |
| `--base-dir` | `-b`      | Sets the base directory for Olares.<br> Defaults to `$HOME/.olares`.                                                                                                                                                                    |
| `--help`     | `-h`      | Displays help information.                                                                                                                                                                                                              |
| `--phase`    |           | Uninstalls Olares from a specific phase and revert to the previous one. <br> For example, `--phase install` removes tasks performed during the "install" phase, reverting the system to the "prepare" stage. <br>Defaults to `install`. |
| `--quiet`    |           | Enables quiet mode (suppress output). <br> Defaults to `false`.                                                                                                                                                                         |
| `--version`  | `-v`      | Specifies the Olares version to uninstall. <br>Use `olares-cli olares info` to check the downloaded version first.                                                                                                                      |

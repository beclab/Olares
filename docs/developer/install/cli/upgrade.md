# `upgrade`

## Synopsis

The `upgrade` command provides a set of tools for upgrading Olares and checking upgrade readiness and compatibility.

```bash
olares-cli upgrade <subcommand> [options]
```

## Subcommands

| Subcommand | Aliases | Description | 
|--|--|--|
| `precheck` | | Prechecks Olares for upgrade. |
| `spec` | `current-spec` | Gets the upgrade spec of the current CLI version. |
| `viable` | | Determines whether upgrade can be directly performed upon a base version. |

## Global options

These options apply to the main `upgrade` command and are inherited by its subcommands where applicable.

| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--base-dir` | `-b` | Sets the base directory for Olares packages. | No | `$HOME/.olares` |
| `--help` | `-h` | Displays help information. | No | N/A |
| `--version` | `-v` | Sets the target Olares version to upgrade to. For example, `1.10.0`, `1.10.0-20241109`. | No | N/A |

## Option for `viable`

| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--base` | `-b` | Base version to check. | No | Current Olares system version |

:::warning Option conflict 
The `-b` shorthand is used by the parent command for `--base-dir`. However, when running `upgrade viable`, `-b` specifically refers to `--base`. 
:::

## Example

```bash
# Check whether the current system can be upgraded directly
sudo olares-cli upgrade viable

# Check upgrade viability from a specific base version
sudo olares-cli upgrade viable --base 1.9.0

# Run pre-upgrade checks
sudo olares-cli upgrade precheck

# View the upgrade spec of the current CLI
sudo olares-cli upgrade spec
```
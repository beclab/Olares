# `delete`

## Synopsis

The `delete` subcommand permanently removes an existing user account from the Olares system. 

**Aliases**: `delete`, `d`, `del`, `rm`, `remove`

```bash
olares-cli user delete <name> [options]
```

## Arguments

| Argument | Description | Required|
|--|--|--|
| `<name>` | Specifies the username of the account to be deleted. <br>It is typically the part before the `@` symbol in an Olares ID. <br>For example, `alice123` for `alice123@olares.com`.| Yes |

## Options
| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--help` | `-h` | Displays help information. | No | N/A |
| `--kubeconfig` | | Specifies the path to a kubeconfig file. | No | N/A |

## Examples

```bash
# Delete a user named alice123
olares-cli user delete alice123
```
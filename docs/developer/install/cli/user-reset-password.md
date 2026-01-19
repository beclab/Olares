# `reset-password`

## Synopsis

The `reset-password` subcommand forcefully resets the login password for a specific user via the authentication provider.

```bash
olares-cli user reset-password <username> [options]
```

## Arguments

| Argument | Description | Required|
|--|--|--|
| `<username>` | Specifies the username of the account to reset. <br>It is typically the part before the `@` symbol in an Olares ID. <br>For example, `alice123` for `alice123@olares.com`.| Yes |

## Options
| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
 `--help` | `-h` | Displays help information. | No | N/A |
| `--kubeconfig` | | Specifies the path to a kubeconfig file. | No | N/A |
| `--password` | `-p` | Specifies the new password for the user. | Yes | N/A |

## Examples

```bash
# Reset password for user named alice123
olares-cli user reset-password alice123 -p "NewSecurePassword456!"
```
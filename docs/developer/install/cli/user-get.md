# `get`

## Synopsis

The `get` subcommand retrieves detailed information about a specific Olares user account. The output can be formatted as a table or JSON. 

```bash
olares-cli user get <name> [options]
```

## Arguments

| Argument | Description | Required|
|--|--|--|
| `<name>` | Specifies the username of the account to retrieve. <br>It is typically the part before the `@` symbol in an Olares ID. <br>For example, `alice123` for `alice123@olares.com`.| Yes |

## Options
| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--help` | `-h` | Displays help information. | No | N/A |
| `--kubeconfig` | | Specifies the path to a kubeconfig file. | No | N/A |
| `--no-headers` | | Disables the header row in the output. | No | N/A |
| `--output` | `-o` | Specifies the output format. Valid values: `table`, `json`. | No | `table` |

## Examples

```bash
# Get details for user named alice123 in default table format
olares-cli user get alice123

# Get details for user named blake123 in JSON format
olares-cli user get blake123 -o json
```
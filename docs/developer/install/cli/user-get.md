# `get`

## Synopsis
:::info
This command typically requires `sudo`.
:::

The `get` subcommand retrieves detailed information about a specific Olares user account. The output can be formatted as a table or JSON. 

```bash
olares-cli user get <name> [options]
```

## Arguments

| Argument | Description | Required|
|--|--|--|
| `<name>` | Specifies the username of the account to retrieve. <br>It is typically the part before the `@` symbol in an Olares ID. <br>For example, `alice123` for `alice123@olares.com`.| **Yes** |

## Options
| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--help` | `-h` | Displays help information. | No | N/A |
| `--kubeconfig` | | Specifies the path to a kubeconfig file. | No | N/A |
| `--no-headers` | | Disables the header row in the output. | No | N/A |
| `--output` | `-o` | Specifies the output format. Valid values: `table`, `json`. | No | `table` |

## Example

```bash
# Get details for user named alice in default table format
sudo olares-cli user get alice

# Get details for user named bob in JSON format
sudo olares-cli user get bob -o json
```
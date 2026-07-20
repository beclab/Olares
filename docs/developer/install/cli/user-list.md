# `list`

## Synopsis

The `list` subcommand displays a list of all registered users in the Olares system. You can sort, filter, and format the output to suit your needs. 

**Aliases**: `list`, `ls`, `l`

```bash
olares-cli user list [options]
```

## Options
| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--help` | `-h` | Displays help information. | No | N/A |
| `--kubeconfig` | | Specifies the path to a kubeconfig file. | No | N/A |
| `--no-headers` | | Disables the header row in the output. | No | N/A |
| `--output` | `-o`| Specifies the output format. Valid values: `table`, `json`.| No | `table` |
| `--reverse` | `-r` | Reverses the sort order of the output. | No | N/A |
| `--sort` | | Sorts the output by a specific field.<br> Valid values: `name`, `role`, `create-time`, `memory`, `cpu` | No | N/A |

## Examples

```bash
# List all users in a table format
olares-cli user list

# List all users in JSON format
olares-cli user list -o json

# List users sorted by role
olares-cli user list --sort role
```
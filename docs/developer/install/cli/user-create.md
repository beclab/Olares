# `create`

## Synopsis

The `create` subcommand creates a new user account in the Olares system. It allows administrators to define initial settings such as the username, password, role permissions, and resource limits. 

**Aliases**: `create`, `add`, `new`

```bash
olares-cli user create <name> [options]
```

## Arguments

| Argument | Description | Required|
|--|--|--|
| `<name>` | Specifies the username for the new account. <br>It is typically the part before the `@` symbol in an Olares ID. <br>For example, `alice123` for `alice123@olares.com`.| Yes |

## Options
| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--cpu-limit` | `-c` | Sets the CPU limit for the user environment. | No | `1` |
| `--description` | | Adds a description for the user account. | No | N/A |
| `--display-name` | | Sets the display name for the user. | No | N/A |
| `--domain` | | Specifies the domain for the Olares ID. | No | Olares system's domain |
| `--help` | `-h` | Displays help information. | No | N/A |
| `--kubeconfig` | | Specifies the path to a kubeconfig file. | No | N/A |
| `--memory-limit` | `-m` | Sets the memory limit for the user environment. | No | `3G` |
| `--password` | `-p` | Sets the initial login password for the user. | No | N/A |
| `--role` | `-r` | Sets the user role.<br>Valid values: `owner`, `admin`, `normal`. | No | `normal` |

## Example

```bash
# Create a basic user with default settings
sudo olares-cli user create alice123

# Create a user with a specified password and role
sudo olares-cli user create blake123 -p "StrongPassword123" -r admin

# Create a user with custom resource limits (2 CPU cores, 4 GB memory)
sudo olares-cli user create carol123 --cpu-limit 2 --memory-limit 4G

# Create a user with display name and description
sudo olares-cli user create david123 --display-name "David Smith" --description "Data platform administrator"
```
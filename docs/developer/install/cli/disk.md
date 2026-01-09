# `disk`

## Synopsis

The `disk` command provides a set of tools to manage storage resources in the Olares system. It is specifically used for managing LVM-based storage configurations.


```bash
olares-cli disk <subcommand>
```

## Subcommands

| Subcommand | Description | 
|--|--|
| `extend` | Extends Olares storage capacity on LVM-based installations. |
| `list-unmounted` | Lists unmounted disks. |

## Flag

| Name | Shorthand | Usage | 
|--|--|--|
| `--help` | `-h` | Displays help information.|

## Example

```bash
# List all disks that are connected but not mounted
sudo olares-cli disk list-unmounted

# Extend Olares storage by adding newly detected unmounted disks
sudo olares-cli disk extend
```
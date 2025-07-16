# `stop`

## Synopsis
The `stop` command is used to stop the components of an installed (or partially installed) Olares system.

```bash
olares-cli stop [option]
```

## Options

| Option             | Shorthand | Usage                                                                                                     | Required | Default |
|--------------------|-----------|-----------------------------------------------------------------------------------------------------------|-------------------------|---------|
| `--check-interval` |           | Specifies the interval between checks for remaining processes during shutdown (e.g., `5s`, `2m`, `3h`).   | No                   | `10s`   |
| `--help`           | `-h`      | Displays help information.                                                                                | No                   | N/A     |
| `--timeout`        |           | Sets the maximum time to wait for a graceful shutdown before using SIGKILL (e.g., `5s`, `2m`, `3h`).       | No                   | `1m`    |

## Example
```bash
# Stop the Olares system
olares-cli stop

# Adjust the timeout for shutdown
olares-cli stop --timeout 2m

# Specify a custom check interval
olares-cli stop --check-interval 5s
```
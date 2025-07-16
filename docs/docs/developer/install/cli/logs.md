# `logs`

## Synopsis
The `logs` command retrieves logs from Olares components and services found on the local machine. It searches for each component listed below, collects logs if the component is found, and skips it if not:

* K3s/Kubelet logs
* Containerd logs
* JuiceFS logs
* Redis logs
* MinIO logs
* etcd logs
* olaresd logs
* Kubernetes pod info and logs
* Kubernetes node info

```bash
olares-cli logs [option]
```

## Options

| Option                 | Shorthand | Usage                                                                                                                                                                                            | Required | Default                        |
|------------------------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|--------------------------------|
| `--components`         |           | Collects logs from specific components (comma-separated).<br/>Supported values: `k3s`, `containerd`, `olaresd`, `kubelet`, `juicefs`, `redis`, `minio`, `etcd`, `NetworkManager`.                  | No                   | All detectable components      |
| `--help`               | `-h`      | Displays help information.                                                                                                                                                                       | No                   | N/A                            |
| `--ignore-kube-errors` |           | Ignores errors from `kubectl` commands (e.g., when Kubernetes API is unreachable) and continues collecting logs.                                                                                 | No                   | `false`                        |
| `--max-lines`          |           | Limits the maximum number of lines for each component's logs to prevent large log files.                                                                                                           | No                   | `3000`                         |
| `--output-dir`         |           | Saves logs to the specified directory. Creates the directory if it does not exist.                                                                                                                 | No                   | `./olares-logs`                |
| `--since`              |           | Fetches logs newer than a specified relative duration (e.g., `5s`, `2m`, `3h`).                                                                                                                    | No                   | `7d`                           |

## Example
```bash
# Collect all logs with default settings
olares-cli logs

# Collect logs for specific components
olares-cli logs --components k3s,redis,minio

# Collect logs for the last 3 hours only
olares-cli logs --since 3h
```
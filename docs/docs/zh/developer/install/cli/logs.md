# `logs`

## 命令说明
`logs` 命令用于获取本机上的 Olares 组件和服务日志。它会检查以下每个组件，如果找到则收集其日志，未找到则跳过：

* K3s/Kubelet 日志
* Containerd 日志
* JuiceFS 日志
* Redis 日志
* MinIO 日志
* etcd 日志
* olaresd 日志
* Kubernetes Pod 信息和日志
* Kubernetes 节点信息

```bash
olares-cli logs [选项]
```

## 选项

| 选项                   | 简写   | 用途                                                                                                                                           | 是否必需 | 默认值                 |
|------------------------|------|----------------------------------------------------------------------------------------------------------------------------------------------|----------|------------------------|
| `--components`         |      | 指定要收集日志的组件（用逗号分隔）。<br/>支持的组件：`k3s`、`containerd`、`olaresd`、`kubelet`、`juicefs`、`redis`、`minio`、`etcd`、`NetworkManager`。 | 否       | 所有可检测到的组件日志 |
| `--help`               | `-h` | 显示命令帮助信息。                                                                                                                                     | 否       | 无                     |
| `--ignore-kube-errors` |      | 忽略 `kubectl` 命令的错误（例如无法连接 Kubernetes API）并继续收集其他日志。                                                                                           | 否       | `false`                |
| `--max-lines`          |      | 限制每个组件日志的最大行数，避免日志文件过大。                                                                                                                   | 否       | `3000` 行              |
| `--output-dir`         |      | 设置日志保存目录。如果目录不存在则自动创建。                                                                                                                 | 否       | `./olares-logs`        |
| `--since`              |      | 设置日志收集的时间范围（例如 `5s`、`2m`、`3h`）。                                                                                                       | 否       | 最近 `7d`（7天）的日志   |

## 使用示例
```bash
# 使用默认设置收集所有日志
olares-cli logs

# 收集指定组件的日志
olares-cli logs --components k3s,redis,minio

# 只收集最近 3 小时的日志
olares-cli logs --since 3h
```
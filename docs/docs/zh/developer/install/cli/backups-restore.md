---
outline: [2, 3]
---
# `restore`
:::warning
必须先运行 `olares-cli backups download` 命令来安装 Restic，否则直接运行此命令将返回错误。
:::
## 命令说明
`restore` 子命令用于从指定的备份仓库和快照中恢复数据到目标目录。

```bash
olares-cli backups restore <存储后端> --path <恢复路径> --repo-name <仓库名称> --snapshot-id <快照ID> [选项]
```

## 通用选项
以下选项适用于所有后端：

| 选项            | 简写   | 用途                   | 是否必需 | 默认值 |
|-----------------|------|------------------------|----------|--------|
| `--help`        | `-h` | 显示命令帮助信息。           | 否       | 无     |
| `--path`        |      | 设置数据恢复的目标目录路径。       | 否       | 无     |
| `--repo-name`   |      | 设置要恢复数据的备份仓库名称。   | 否       | 无     |
| `--snapshot-id` |      | 设置要恢复的快照 ID。        | 否       | 无     |

## 存储后端配置选项

### 腾讯云对象存储（`cos`）选项

| 选项                    | 简写 | 用途                                                                         | 是否必需 | 默认值   |
|-------------------------|----|------------------------------------------------------------------------------|----------|----------|
| `--access-key`          |    | 设置腾讯云 COS 的访问密钥。                                                            | 否       | 无       |
| `--endpoint`            |    | 设置腾讯云 COS 的终端节点，格式如：`https://cos.{region}.myqcloud.com/{bucket}/{prefix}`。 | 否       | 无       |
| `--limit-download-rate` |    | 设置下载速度的最大值，单位为 KiB/s。                                                       | 否       | 不限速   |
| `--secret-access-key`   |    | 设置腾讯云 COS 的密钥。                                                          | 否       | 无       |

### 本地文件系统（`fs`）选项

| 选项          | 简写 | 用途                 | 是否必需 | 默认值 |
|---------------|----|----------------------|----------|--------|
| `--endpoint`  |    | 设置存储备份的本地目录路径。 | 否       | 无     |
| `--olares-id` |    | 设置 Olares ID。      | 否       | 无     |

### Amazon S3 选项（`s3`）

| 选项                    | 简写 | 用途                                                                         | 是否必需 | 默认值   |
|-------------------------|----|------------------------------------------------------------------------------|----------|----------|
| `--access-key`          |    | 设置 Amazon S3 的访问密钥。                                                        | 否       | 无       |
| `--endpoint`            |    | 设置 Amazon S3 的终端节点，格式如：`https://{bucket}.{region}.amazonaws.com/{prefix}`。 | 否       | 无       |
| `--limit-download-rate` |    | 设置下载速度的最大值，单位为 KiB/s。                                                       | 否       | 不限速   |
| `--secret-access-key`   |    | 设置 Amazon S3 的密钥。                                                          | 否       | 无       |

### Olares Space 选项（`space`）

| 选项                          | 简写 | 用途                                                                  | 是否必需 | 默认值   |
|-------------------------------|----|---------------------------------------------------------------------|----------|----------|
| `--access-token` <sup>1</sup> |    | 设置 Olares Space 的访问令牌。                                              | 否       | 无       |
| `--cloud-api-mirror`          |    | 设置云 API 镜像地址。                                                       | 否       | 无       |
| `--cloud-name`                |    | 设置 Olares Space 实例的云名称。<br/> 可通过 [`region`](./backups-region.md) 子命令获取。       | 否       | 无       |
| `--cluster-id` <sup>2</sup>   |    | 设置用于存储备份的集群 ID。                                                     | 否       | 无       |
| `--limit-download-rate`       |    | 设置下载速度的最大值，单位为 KiB/s。                                            | 否       | 不限速   |
| `--olares-did` <sup>1</sup>   |    | 设置 Olares DID。                                                      | 否       | 无       |
| `--region-id`                 |    | 设置 Olares Space 实例的区域 ID。<br/> 可通过 [`region`](./backups-region.md) 子命令获取。 | 否       | 无       |

1. 要获取访问令牌和 Olares DID，请在登录 Olares Space 后检查页面网络请求的负载。`token` 字段对应访问令牌，`userid` 字段对应 Olares DID。

2. 要获取集群 ID，请运行以下命令：
   ```bash
   kubectl get terminus -o jsonpath='{.items[*].metadata.labels.bytetrade\.io/cluster-id}'
   ```
## 使用示例
```bash
# 从腾讯云对象存储恢复数据
olares-cli backups restore cos --path /data_restore --repo-name my_repo \
  --snapshot-id snapshot_12345 \
  --access-key YOUR_KEY \
  --secret-access-key YOUR_SECRET \
  --endpoint https://cos.region.myqcloud.com/bucket/prefix
  
# 从本地文件系统恢复数据
olares-cli backups restore fs --path /data_restore --repo-name my_repo \
  --snapshot-id snapshot_12345 --endpoint /backup_repo
```
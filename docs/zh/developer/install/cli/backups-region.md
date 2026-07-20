# `region`

## 命令说明
`region` 子命令用于获取云名称和区域 ID。该命令专门用于存储后端为 Olares Space 的场景。
```bash
olares-cli backups region space [选项]
```

## 选项

| 名称	                           | 简写   | 用途                     |
|-------------------------------|------|------------------------|
| `--access-token` <sup>1</sup> |      | 设置 Olares Space 的访问令牌。 |
| `--cloud-api-mirror`          |      | 设置云 API 镜像地址。          |
| `--help`                      | `-h` | 显示命令帮助信息。              |
| `--olares-did` <sup>1</sup>   |      | 设置 Olares DID。         |

1. 要获取访问令牌和 Olares DID，请在登录 Olares Space 后检查页面网络请求的负载。`token` 字段对应访问令牌，`userid` 字段对应 Olares DID。

## 使用示例
```bash
# 查询云名称和区域 ID
olares-cli backups region space \
--access-token YOUR_ACCESS_TOKEN \
--cloud-api-mirror https://api-mirror.example.com \
--olares-did did:xyz123
```
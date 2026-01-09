# `create`

## 命令说明

`create`子命令用于在 Olares 中创建一个新的用户。管理员可以在创建时设置用户名、初始密码、角色权限以及资源限制等初始配置。

**别名**：`create`、`add`、`new`

```bash
olares-cli user create <用户名> [选项]
```

## 参数

| 参数 | 说明 | 是否必需|
|--|--|--|
| `<用户名>` | 指定新用户的用户名。通常为 Olares ID 中`@`符号之前的部分。<br>例如`alice123@olares.com`中的`alice123`。| 是 |

## 选项
| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--cpu-limit` | `-c` | 设置用户环境的 CPU 限制。 | 否 | `1` |
| `--description` | | 为用户账号添加描述信息。 | 否 | 无 |
| `--display-name` | | 设置用户的显示名称。 | 否 | 无 |
| `--domain` | | 指定 Olares ID 使用的域名。 | 否 | 系统默认域名 |
| `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--kubeconfig` | | 指定 kubeconfig 文件路径。 | 否 | 无 |
| `--memory-limit` | `-m` | 设置用户环境的内存限制。 | 否 | `3G` |
| `--password` | `-p` | 设置用户的初始登录密码。 | 否 | 无 |
| `--role` | `-r` | 设置用户角色。<br>可选值：`owner`、`admin`、`normal`。 | 否 | `normal` |

## 使用示例

```bash
# 使用默认设置创建用户
sudo olares-cli user create alice123

# 创建用户并指定密码和角色
sudo olares-cli user create blake123 -p "StrongPassword123" -r admin

# 创建用户并设置资源限制（2 个 CPU 核心，4 GB 内存）
sudo olares-cli user create carol123 --cpu-limit 2 --memory-limit 4G

# 创建用户并设置显示名称和描述
sudo olares-cli user create david123 --display-name "David Smith" --description "Data platform administrator"
```
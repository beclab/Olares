# `reset-password`

## 命令说明

`reset-password`子命令用于通过认证服务强制重置指定用户的登录密码。

```bash
olares-cli user reset-password <用户名> [选项]
```

## 参数

| 参数 | 说明 | 是否必需 |
|--|--|--|
| `<用户名>` | 指定要重置密码的用户名。通常为 Olares ID 中`@`符号之前的<br>部分。例如`alice123@olares.com`中的`alice123`。| 是 |

## 选项
| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
 `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--kubeconfig` | | 指定 kubeconfig 文件路径。 | 否 | 无 |
| `--password` | `-p` | 指定为该用户设置的新密码。 | 是 | 无 |

## 使用示例

```bash
# 重置用户 alice123 的密码
sudo olares-cli user reset-password alice123 -p "NewSecurePassword456!"
```
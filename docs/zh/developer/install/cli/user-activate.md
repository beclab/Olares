# `activate`

## 命令说明

:::info 执行权限
该命令通常需要使用管理员权限（`sudo`）执行。
:::

`activate`子命令用于激活已有的 Olares 账户，你需要至少提供 Olares ID、登录密码和助记词短语才能完成激活。

```bash
olares-cli user activate <Olares ID> [选项]
```
## 参数

| 参数 | 描述 | 是否必需 |
|--|--|--|
| `<Olares ID>` | 指定待激活的 Olares 用户账户的唯一标识符。<br>格式类似电子邮件地址（例如：`alice123@olares.com`）| **是** |

## 选项

| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--bfl` | | 指定聚合后端接口（BFL）服务地址。<br>例如：`https://example.com` | 否 | `http://127.0.0.1:30180` |
| `--enable-tunnel` | | 启用或关闭隧道模式进行激活。 | 否 | `false` |
| `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--host` | | 指定快速反向代理（FRP）主机地址。<br>仅当`--enable-tunnel`设置为`true`时使用。 | 否 | 无 |
| `--jws` | | 指定指定 FRP 的 JWS（JSON Web Signature）令牌。<br>仅当`--enable-tunnel`设置为`true`时使用。| 否 | 无 |
| `--language` | | 设置系统语言。| 否 | `en-US` |
| `--location` | | 设置系统时区位置。 | 否 | `Asia/Shanghai` |
| `--mnemonic` | | 指定用于激活的 12 词助记词短语。 | **是** | 无 |
| `--password` | `-p` | 指定待激活 Olares 账户的登录密码。 | **是** | 无 |
| `--reset-password` | | 指定用于密码重置的新密码。<br>仅在执行密码重置时需要。| 否 | 无 |
| `--vault` | | 指定 Vault 服务的 URL。<br>例如：`https://example.com` | 否 | `http://127.0.0.1:30181` |

## 使用示例

```bash
# 激活 Olares 账户
sudo olares-cli user activate alice@olares.cn -p "HerPassWord"  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# 启用隧道模式激活 Olares 账户
sudo olares-cli user activate david@olares.cn -p "HisPassWord"  --enable-tunnel --host "frp-gateway.olares.com"  --jws "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.demo.signature"  --bfl http://127.0.0.1:30180 --vault http://127.0.0.1:30180/server  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# 使用指定的语言和时区设置，激活 Olares 账户
sudo olares-cli user activate carol@olares.cn -p "AnotherPassWord"  --mnemonic "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"  --language "cn-ZH" --location "Asia/Shanghai"
```
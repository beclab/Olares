# `user activate`

## 命令说明

`user activate`命令用于激活 Olares 账户。

```bash
olares-cli user activate <Olares ID> [选项]
```
## 参数

| 参数 | 描述 | 是否必需 |
|--|--|--|
| `<Olares ID>` | Olares 生态系统中的唯一标识符，<br>格式类似电子邮件地址（例如：`alice123@olares.com`）| **是** |

## 选项

| 选项 | 简写 | 用途 | 是否必需 | 默认值 |
|--|--|--|--|--|
| `--bfl` | | 指定聚合后端接口（BFL）服务地址。<br>例如：`https://example.com` | 否 | `http://127.0.0.1:30180` |
| `--enable-tunnel` | | 启用隧道模式进行激活。 | 否 | `false` |
| `--help` | `-h` | 显示帮助信息。 | 否 | 无 |
| `--host` | | 指定快速反向代理（FRP）主机地址。<br>仅当`--enable-tunnel`设置为`true`时使用。 | 否 | 无 |
| `--jws` | | 指定指定 FRP 的 JWS（JSON Web Signature）令牌。<br>仅当`--enable-tunnel`设置为`true`时使用。| 否 | 无 |
| `--language` | | 设置系统语言。| 否 | `en-US` |
| `--location` | | 设置系统时区位置。 | 否 | `Asia/Shanghai` |
| `--mnemonic` | | 指定用于激活的 12 个助记词。 | **是** | 无 |
| `--password` | `-p` | 指定用于激活的 Olares 登录密码。 | **是** | 无 |
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
# `user activate`

## Synopsis

The `user activate` command is used to activate an existing Olares user account.

```bash
olares-cli user activate <Olares ID> [options]
```

## Arguments

| Argument | Description | Required|
|--|--|--|
| `<Olares ID>` | The unique user identifier within the Olares ecosystem, <br>similar to an email address(e.g., `alice123@olares.com`).| **Yes** |

## Options
| Option | Shorthand | Usage | Required | Default |
|--|--|--|--|--|
| `--bfl` | | Specifies the Backend For Launcher (BFL) service URL (e.g., `https://example.com`). | No | `http://127.0.0.1:30180` |
| `--enable-tunnel` | | Enables tunnel mode for activation. | No | `false` |
| `--help` | `-h` | Displays help information. | No | N/A |
| `--host` | | Specifies the Fast Reverse Proxy (FRP) host. <br>Only used when the `--enable-tunnel` option is set to `true`. | No | N/A |
| `--jws` | | Specifies the FRP JWS token.<br>Only used when the `--enable-tunnel` option is set to `true`.| No | N/A |
| `--language` | | Sets the system language. | No | `en-US` |
| `--location` | | Sets the timezone location. | No | `Asia/Shanghai` |
| `--mnemonic` | | Specifies the 12-word mnemonic phrase required for activation. | **Yes** | N/A |
| `--password` | `-p` | Specifies the Olares login password for authentication.	 | **Yes** | N/A |
| `--vault` | | Specifies the Vault service URL (e.g., `https://example.com`). | No | `http://127.0.0.1:30181` |

## Example

```bash
# Activate an Olares user account
sudo olares-cli user activate alice@olares.com -p "HerPassWord"  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# Activate an Olares user account with tunnel mode enabled
sudo olares-cli user activate david@olares.com -p "HisPassWord"  --enable-tunnel --host "frp-gateway.olares.com"  --jws "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.demo.signature"  --bfl http://127.0.0.1:30180 --vault http://127.0.0.1:30180/server  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# Activate an Olares user account with specific language and timezone settings
sudo olares-cli user activate carol@olares.com -p "AnotherPassWord"  --mnemonic "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"  --language "en-US" --location "America/New_York"
```
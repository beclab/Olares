# `activate`

## Synopsis

The `activate` subcommand activates an existing Olares account. It requires the user's Olares ID, password, and 12-word mnemonic phrase to complete the activation. 


```bash
olares-cli user activate <Olares ID> [options]
```

## Arguments

| Argument | Description | Required|
|--|--|--|
| `<Olares ID>` | Specifies the unique identifier for the Olares user account <br>to be activated. <br>Similar to an email address like `alice123@olares.com`.| Yes |

## Options
| Option | Short| Usage | Required | Default |
|--|--|--|--|--|
| `--bfl` | | Specifies the Backend For Launcher (BFL) service URL, such as `https://example.com`. | No | `http://127.0.0.1:30180` |
| `--enable-tunnel` | | Enables or disables tunnel mode for activation. | No | `false` |
| `--help` | `-h` | Displays help information. | No | N/A |
| `--host` | | Specifies the Fast Reverse Proxy (FRP) host. Only used when `--enable-tunnel` option is set to `true`. | No | N/A |
| `--jws` | | Specifies the FRP JWS token. Only used when `--enable-tunnel` option is set to `true`.| No | N/A |
| `--language` | | Sets the system language. | No | `en-US` |
| `--location` | | Sets the timezone location. | No | `Asia/Shanghai` |
| `--mnemonic` | | Specifies the 12-word mnemonic phrase required for activation. | Yes | N/A |
| `--password` | `-p` | Specifies the Olares login password for authentication.	 | Yes | N/A |
| `--reset-password` | | Specifies the new password to set during password reset. This option is required only when performing a password reset. | No | N/A |
| `--vault` | | Specifies the Vault service URL, such as `https://example.com`. | No | `http://127.0.0.1:30180` |

## Examples

```bash
# Activate an Olares user account
olares-cli user activate alic123e@olares.com -p "HerPassWord"  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# Activate an Olares user account with tunnel mode enabled
olares-cli user activate david456@olares.com -p "HisPassWord"  --enable-tunnel --host "frp-gateway.olares.com"  --jws "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.demo.signature"  --bfl http://127.0.0.1:30180 --vault http://127.0.0.1:30180/server  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# Activate an Olares user account with specific language and timezone settings
olares-cli user activate carol789@olares.com -p "AnotherPassWord"  --mnemonic "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"  --language "en-US" --location "America/New_York"
```
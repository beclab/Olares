# `user activate`

## Synopsis

The `user activate` command is used to activate an existing Olares user account using the specified Olares ID, mnemonic phrase, and password.

```bash
olares-cli user activate {Olares ID} [options]
```

## Options

<table style="width:100%; table-layout:fixed; border-collapse:collapse;">
  <colgroup>
    <col style="width: 16%;" />  
    <col style="width: 8%;" />   
    <col style="width: 35%;" />   
    <col style="width: 10%;" />   
    <col style="width: 30%;" />  
  </colgroup>
  <thead>
    <tr>
      <th>Option</th>
      <th>Shorthand</th>
      <th>Usage</th>
      <th>Required</th>
      <th>Default</th>
    </tr>
  </thead>
    <tbody>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--bfl</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Specifies the Bfl (Backend For Launcher) service URL (e.g., <code>https://example.com</code>).</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;"><code>http://127.0.0.1:30180</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--enable-tunnel</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Enables tunnel mode for activation.</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;"><code>false</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--help</code></td>
      <td style="text-align:left; word-break:break-word;"><code>-h</code></td>
      <td style="text-align:left; word-break:break-word;">Displays help information.</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;">N/A</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--host</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Specifies the FRP (Fast Reverse Proxy) host. <br>Only used when tunnel mode is enabled.</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;">N/A</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--jws</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Specifies the FRP JWS token. <br>Only used when tunnel mode is enabled.</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;">N/A</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--language</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Sets the system language.</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;"><code>en-US</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--location</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Sets the timezone location.</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;"><code>Asia/Shanghai</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--mnemonic</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Specifies the 12-word mnemonic phrase required for activation.</td>
      <td style="text-align:left; word-break:break-word;">Yes</td>
      <td style="text-align:left; word-break:break-word;">N/A</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--password</code></td>
      <td style="text-align:left; word-break:break-word;"><code>-p</code></td>
      <td style="text-align:left; word-break:break-word;">Specifies the Olares login password for authentication.</td>
      <td style="text-align:left; word-break:break-word;">Yes</td>
      <td style="text-align:left; word-break:break-word;">N/A</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--vault</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">Specifies the Vault service URL (e.g., <code>https://example.com</code>).</td>
      <td style="text-align:left; word-break:break-word;">No</td>
      <td style="text-align:left; word-break:break-word;"><code>http://127.0.0.1:30181</code></td>
    </tr>
  </tbody>
</table>

## Example

```bash
# Activate an Olares user account
sudo olares-cli user activate alice@olares.com -p "HerPassWord"  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# Activate an Olares user account with tunnel mode enabled
sudo olares-cli user activate david@olares.com -p "HisPassWord"  --enable-tunnel --host "frp-gateway.olares.com"  --jws "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.demo.signature"  --bfl http://127.0.0.1:30180 --vault http://127.0.0.1:30180/server  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# Activate an Olares user account with specific language and timezone settings
sudo olares-cli user activate carol@olares.com -p "AnotherPassWord"  --mnemonic "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"  --language "en-US" --location "America/New_York"
```
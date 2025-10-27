# `user activate`

## 命令说明

`user activate`命令用于使用指定的 Olares ID、助记词和密码来激活 Olares 账户。

```bash
olares-cli user activate {Olares ID} [选项]
```

## 选项

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
      <th>选项</th>
      <th>简写</th>
      <th>用途</th>
      <th>是否必需</th>
      <th>默认值</th>
    </tr>
  </thead>
    <tbody>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--bfl</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">指定 Bfl （聚合后端接口）服务 URL。<br>例如：<code>https://example.com</code></td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;"><code>http://127.0.0.1:30180</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--enable-tunnel</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">启用隧道模式进行激活</td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;"><code>false</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--help</code></td>
      <td style="text-align:left; word-break:break-word;"><code>-h</code></td>
      <td style="text-align:left; word-break:break-word;">显示帮助信息。</td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;">无</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--host</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">指定 FRP（Fast Reverse Proxy，快速反向代理）的主机地址。<br>
仅在启用隧道模式时使用。</td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;">无</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--jws</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">指定 FRP 的 JWS（JSON Web Signature）令牌。
<br>仅在启用隧道模式时使用。</td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;">无</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--language</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">设置系统语言。</td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;"><code>en-US</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--location</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">设置系统时区位置。</td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;"><code>Asia/Shanghai</code></td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--mnemonic</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">指定用于激活的 12 个助记词。</td>
      <td style="text-align:left; word-break:break-word;">是</td>
      <td style="text-align:left; word-break:break-word;">无</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--password</code></td>
      <td style="text-align:left; word-break:break-word;"><code>-p</code></td>
      <td style="text-align:left; word-break:break-word;">指定用于激活的 Olares 登录密码。</td>
      <td style="text-align:left; word-break:break-word;">是</td>
      <td style="text-align:left; word-break:break-word;">无</td>
    </tr>
    <tr>
      <td style="text-align:left; word-break:break-word;"><code>--vault</code></td>
      <td style="text-align:left; word-break:break-word;"></td>
      <td style="text-align:left; word-break:break-word;">指定 Vault 服务的 URL。<br>例如：<code>https://example.com</code></td>
      <td style="text-align:left; word-break:break-word;">否</td>
      <td style="text-align:left; word-break:break-word;"><code>http://127.0.0.1:30181</code></td>
    </tr>
  </tbody>
</table>

## 使用示例

```bash
# 激活 Olares 账户
sudo olares-cli user activate alice@olares.cn -p "HerPassWord"  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# 启用隧道模式激活 Olares 账户
sudo olares-cli user activate david@olares.cn -p "HisPassWord"  --enable-tunnel --host "frp-gateway.olares.com"  --jws "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.demo.signature"  --bfl http://127.0.0.1:30180 --vault http://127.0.0.1:30180/server  --mnemonic "apple banana cherry door eagle forest grape house island jacket kite lemon"

# 使用指定的语言和时区设置，激活 Olares 账户
sudo olares-cli user activate carol@olares.cn -p "AnotherPassWord"  --mnemonic "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu"  --language "cn-ZH" --location "Asia/Shanghai"
```
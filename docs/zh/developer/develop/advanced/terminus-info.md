# olares-info

olares-info 是一个无需鉴权的接口，我们将一些系统可以对外展示的信息放在这里。可以理解为它是房子的门牌。

## 外部调用

```
https://<username>.olares.com/api/olares-info
```

## 字段含义

```json
interface OlaresInfo {
  olaresId: string;
  wizardStatus: string;
  enable_reverse_proxy: boolean;
  tailScaleEnable: boolean;
  osVersion: string;
  avatar: string;
  loginBackground: string;
  id: string;
}
```

## 字段含义

<table style="width:100%; table-layout:fixed; border-collapse:collapse;">
  <colgroup>
    <col style="width:25%;">
    <col style="width:75%;">
  </colgroup>
  <thead>
    <tr>
      <th style="text-align:left;">Field</th>
      <th style="text-align:left;">Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td style="text-align:left;">olaresId	</td>
      <td style="text-align:left; white-space:normal; word-break:break-word;">用户的 Olares ID。格式是：<code>username@domain.com</code>.</td>
    </tr>
    <tr>
      <td>wizardStatus</td>
      <td style="white-space:normal; word-break:break-word;">Olares 的激活状态，有以下状态：<br><code>wait_activate_vault</code>，<code>vault_activating</code>，<code>vault_activate_failed</code>，<code>wait_activate_system</code>，<code>system_activating</code>，<code>system_activate_failed</code>，<code>wait_activate_network</code>，<code>network_activating</code>，<code>network_activate_failed</code>，<code>wait_reset_password</code>，<code>completed</code>。<br>当状态为 <code>completed</code> 时，代表用户激活成功。在用户激活成功前，不推荐第三方程序进行太多业务相关的逻辑。
</td>
    </tr>
    <tr>
      <td>enable_reverse_proxy</td>
      <td>用户是否启用了反向代理</td>
    </tr>
    <tr>
      <td>tailScaleEnable	</td>
      <td>用户是否激活了 TailScale。如果激活了私有入口，只能通过 VPN 访问。<br>用途：LarePass 在连接 Olares 时，不根据这个变量决定是否增加 local 访问。</td>
    </tr>
    <tr>
      <td>osVersion	</td>
      <td>Olares 的系统版本</td>
    </tr>
    <tr>
      <td>avatar</td>
      <td>用户的头像</td>
    </tr>
    <tr>
      <td>loginBackground</td>
      <td>登录界面的背景图</td>
    </tr>
    <tr>
      <td>id</td>
      <td>用户在每次激活时，都会生成一个新的唯一 ID</td>
    </tr>
  </tbody>
</table>
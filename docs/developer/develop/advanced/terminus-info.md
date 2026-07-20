# olares-info

olares-info is an API without authentication. It displays publicly available system information. You can think of it as a house number sign.

## API Call

```
https://<username>.olares.com/api/olares-info
```

## Data Structure

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

## API Field Definitions

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
      <td style="text-align:left; white-space:normal; word-break:break-word;">The user's Olares ID follows a format like <code>username@domain.com</code>.</td>
    </tr>
    <tr>
      <td>wizardStatus</td>
      <td style="white-space:normal; word-break:break-word;">Activation status of Olares, possible statuses includes: <code>wait_activate_vault</code>, <code>vault_activating</code>, <code>vault_activate_failed</code>, <code>wait_activate_system</code>, <code>system_activating</code>, <code>system_activate_failed</code>, <code>wait_activate_network</code>, <code>network_activating</code>, <code>network_activate_failed</code>, <code>wait_reset_password</code>, <code>completed</code>. <br>When the status displays <code>completed</code>, it indicates that the system has been successfully activated. We advise against third-party programs executing excessive business-related logic before the system is fully activated.
</td>
    </tr>
    <tr>
      <td>enable_reverse_proxy</td>
      <td>Whether a reverse proxy is enabled.</td>
    </tr>
    <tr>
      <td>tailScaleEnable	</td>
      <td>Whether the TailScale is activated. If so, all private entrances can only be accessed through the VPN. <br>Note: This field does not affect whether LarePass uses local access when connecting to Olares.</td>
    </tr>
    <tr>
      <td>osVersion	</td>
      <td>Olares version</td>
    </tr>
    <tr>
      <td>avatar</td>
      <td>User's Avatar</td>
    </tr>
    <tr>
      <td>loginBackground</td>
      <td>Background image of the login interface</td>
    </tr>
    <tr>
      <td>id</td>
      <td>Every time the user activates Olares, a new unique ID is generated.</td>
    </tr>
  </tbody>
</table>
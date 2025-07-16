# terminus-info

terminus-info is an API without authentication. It displays publicly available system information. You can think of it as a house number sign.

## API Call

```
https://<username>.olares.com/api/terminus-info
```

## Data Structure

```json
interface TerminusInfo {
  terminusName: string;
  wizardStatus: string;
  selfhosted: boolean;
  tailScaleEnable: boolean;
  osVersion: string;
  avatar: string;
  loginBackground: string;
  terminusId: string;
}
```

## API Field Definitions

| Field           | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
|-----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| terminusName    | The user's Olares ID follows a format like `username@domain.com`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| wizardStatus    | Activation status of Olares, possible statuses includes: `wait_activate_vault`, `vault_activating`, `vault_activate_failed`, `wait_activate_system`, `system_activating`, `system_activate_failed`, `wait_activate_network`, `network_activating`, `network_activate_failed`, `wait_reset_password`, `completed`. When the status displays `completed`, it indicates that the system has been successfully activated. We advise against third-party programs executing excessive business-related logic before the system is fully activated. |
| selfhosted      | Whether the Olares is running on Olares Space                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| tailScaleEnable | Whether the TailScale is activated. If so, all private entrances can only be accessed through the VPN. <br> Note: This field does not affect whether LarePass uses local access when connecting to Olares.                                                                                                                                                                                                                                                                                                                                    |
| osVersion       | Olares version                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| avatar          | User's Avatar                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| loginBackground | Background image of the login interface                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| olaresId        | Every time the user activates Olares, a new unique ID is generated.                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |

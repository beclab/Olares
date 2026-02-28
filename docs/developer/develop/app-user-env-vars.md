---
outline: [2, 3]
description: User environment variables are user-level personalized settings. Apps must reference them via `envs.valueFrom` and map them to `.Values.olaresEnv`.
---

# User environment variables

User environment variables are personalized settings for each Olares user and are managed by the current user. In a cluster with multiple users, each user’s environment variables are independent.

:::info
User environment variables are a "variable pool". Apps must map them to their own `envName` via `envs.valueFrom`, and then use them in templates via `.Values.olaresEnv.<envName>`.
:::

## How to use

1. Declare the mapping in `OlaresManifest.yaml`. Under `envs`, declare an app variable and use `valueFrom` to reference the user environment variable.

    **Example**:

    ```yaml
    olaresManifest.version: "0.10.0"
    olaresManifest.type: app

    envs:
      - envName: USER_TIMEZONE
        valueFrom:
          envName: OLARES_USER_TIMEZONE
    ```
2. In your Helm template where the variable is needed, reference it using the `.Values.olaresEnv` path.

    **Example**:
    ```yaml
    value: "{{ .Values.olaresEnv.USER_TIMEZONE }}"
    ```
3. When the app is deployed via App Service, the system automatically retrieves the current user's configuration and injects it into `values.yaml`.

    **Example output**:

    ```yaml
    olaresEnv:
      USER_TIMEZONE: "Asia/Shanghai"
    ```
## User environment variables reference

### User info

#### OLARES_USER_EMAIL

- Type: `string`
- Editable: Yes

#### OLARES_USER_USERNAME

- Type: `string`
- Editable: Yes

#### OLARES_USER_PASSWORD

- Type: `password`
- Editable: Yes

#### OLARES_USER_TIMEZONE

- Type: `string`
- Editable: Yes

### SMTP settings

#### OLARES_USER_SMTP_ENABLED
Whether to enable SMTP.
- Type: `bool`
- Editable: Yes

#### OLARES_USER_SMTP_SERVER
SMTP server domain.
- Type: `domain`
- Editable: Yes

#### OLARES_USER_SMTP_PORT
SMTP server port, typically "465" or "587".
- Type: `int`
- Editable: Yes

#### OLARES_USER_SMTP_USERNAME
SMTP username.
- Type: `string`
- Editable: Yes

#### OLARES_USER_SMTP_PASSWORD
SMTP password or authorization code.
- Type: `password`
- Editable: Yes

#### OLARES_USER_SMTP_FROM_ADDRESS
Sender email address.
- Type: `email`
- Editable: Yes

#### OLARES_USER_SMTP_SECURE
Whether to use a secure protocol.
- Type: `bool`
- Default: `true`
- Editable: Yes

#### OLARES_USER_SMTP_USE_TLS
Secure protocol option.
- Type: `bool`
- Editable: Yes

#### OLARES_USER_SMTP_USE_SSL
Secure protocol option.
- Type: `bool`
- Editable: Yes

#### OLARES_USER_SMTP_SECURITY_PROTOCOLS
Security protocol. Allowed values: `tls`, `ssl`, `starttls`, `none`
- Type: `string`
- Editable: Yes

### Mirror/Proxy endpoints

#### OLARES_USER_HUGGINGFACE_SERVICE

- Type: `url`
- Default: `"https://huggingface.co/"`
- Editable: Yes

#### OLARES_USER_HUGGINGFACE_TOKEN

- Type: `string`
- Editable: Yes

#### OLARES_USER_PYPI_SERVICE

- Type: `url`
- Default: `"https://pypi.org/simple/"`
- Editable: Yes

#### OLARES_USER_GITHUB_SERVICE

- Type: `url`
- Default: `"https://github.com/"`
- Editable: Yes

#### OLARES_USER_GITHUB_TOKEN

- Type: `string`
- Editable: Yes

### API keys

#### OLARES_USER_OPENAI_APIKEY

- Type: `password`
- Editable: Yes

#### OLARES_USER_CUSTOM_OPENAI_SERVICE

- Type: `url`
- Editable: Yes

#### OLARES_USER_CUSTOM_OPENAI_APIKEY

- Type: `password`
- Editable: Yes

## Full structure example

```yaml
userEnvs:
# User info
  - envName: OLARES_USER_EMAIL
    type: string
    editable: true
  - envName: OLARES_USER_USERNAME
    type: string
    editable: true
  - envName: OLARES_USER_PASSWORD
    type: password
    editable: true
  - envName: OLARES_USER_TIMEZONE
    type: string
    editable: true    

# SMTP settings
    
  - envName: OLARES_USER_SMTP_ENABLED
    type: bool
    editable: true
    
  - envName: OLARES_USER_SMTP_SERVER
    type: domain 
    editable: true
    
  - envName: OLARES_USER_SMTP_PORT
    type: int
    editable: true
    
  - envName: OLARES_USER_SMTP_USERNAME
    type: string
    editable: true
    
  - envName: OLARES_USER_SMTP_PASSWORD
    type: password
    editable: true
    
  - envName: OLARES_USER_SMTP_FROM_ADDRESS
    type: email
    editable: true    
   
  - envName: OLARES_USER_SMTP_SECURE
    type: bool
    editable: true
    default: "true"   
    
  - envName: OLARES_USER_SMTP_USE_TLS
    type: bool
    editable: true
  - envName: OLARES_USER_SMTP_USE_SSL
    type: bool
    editable: true
    
  - envName: OLARES_USER_SMTP_SECURITY_PROTOCOLS
    type: string
    editable: true        
    
# Mirror/Proxy endpoints
  - envName: OLARES_USER_HUGGINGFACE_SERVICE
    type: url
    editable: true
    default: "https://huggingface.co/"
  - envName: OLARES_USER_HUGGINGFACE_TOKEN
    type: string
    editable: true    
  - envName: OLARES_USER_PYPI_SERVICE
    type: url
    editable: true
    default: "https://pypi.org/simple/"
  - envName: OLARES_USER_GITHUB_SERVICE
    type: url
    editable: true
    default: "https://github.com/"
  - envName: OLARES_USER_GITHUB_TOKEN
    type: string
    editable: true

# API keys
  - envName: OLARES_USER_OPENAI_APIKEY
    type: password
    editable: true
  - envName: OLARES_USER_CUSTOM_OPENAI_SERVICE
    type: url
    editable: true    
  - envName: OLARES_USER_CUSTOM_OPENAI_APIKEY
    type: password
    editable: true     
```
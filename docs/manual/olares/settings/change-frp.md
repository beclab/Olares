---
description: Learn how to change the reverse proxy option in Olares Settings to expose internal services securely.
---
# Change reverse proxy

A reverse proxy acts as a secure gateway between your Olares and the open web, enabling you to expose local services to the public internet securely without needing a public IP. For users who do not own a public IP address, Olares offers three reverse options to facilitate external access to Olares applications and services:

- **Cloudflare Tunnel** – Recommended for most users worldwide.

- **Olares Tunnel** – Optimized for users in mainland China with different regional IDC options. 

- **Self-built FRP** – Ideal for users with their own FRP servers.

## Change your reverse proxy option

1. Open Settings, then navigate to **Network** > **Reverse Proxy**.
2. Choose your preferred reverse proxy option. If you select Self-built FRP, you’ll need to provide the server address, port, and authentication method.

3. Click **Save** to apply your changes.

:::warning Change with caution
- Olares sets a default reverse proxy option during installation. Changing this setting may affect connectivity—proceed with caution.
- Switching your reverse proxy between Cloudflare Tunnel and Olares Tunnel will also change your data plan and benefits, as they are separate cloud network services with different pricing and features. 
:::

---
description: Learn using Terminal in Olares to check Olares component status, export logs, and update system settings. 
---
# Access Terminal

The Terminal module in Control Hub provides one-click root access to the Olares host terminal. This is equivalent to connecting via SSH, but without needing to manually enter your account and password.

![Terminal](/images/manual/olares/controlhub-terminal.png#bordered)

You can use the terminal to:

- Perform debugging and troubleshooting directly on the host.
- View logs in real-time or review historical system logs.
- Modify system configurations at the OS level.

:::tip Note
The terminal connects to the underlying Linux system of the Olares host. Operations performed here affect the host environment directly, so please proceed with caution.
:::

## Command examples

```bash
# View the status of GPU drivers and services
olares-cli gpu status

# Collect all logs with default settings
olares-cli logs

# Collect logs for specific components
olares-cli logs --components k3s,redis,minio

# Check whether the NVIDIA GPU is working normally
nvidia-smi

# Check whether the system recognizes the NVIDIA GPU hardware
lspci | grep -i vga | grep -i nvidia
```

For more command usages, refer to [Olares CLI doc](../../../developer/install/cli/olares-cli.md).

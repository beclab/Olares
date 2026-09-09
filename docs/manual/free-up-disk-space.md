---
description: Review disk usage, delete unused AI models and application images, and verify recovered space on Olares.
---
# Free up disk space

If you see errors such as "Authentication failed, disk space is full", or app installations fail because storage is exhausted, review disk usage and clean up content you no longer need.

## Prerequisites

- Administrator access to Olares.
- Terminal access to Olares, for removing unused images. See [Connect to Olares via SSH](/manual/access-olares-terminal).

## Review disk usage

1. Open Dashboard and check disk usage for the system and each application. For details, see [Check system and app resource usage](olares/resources-usage.md).
2. In **Files**, review large files and folders you can move to external storage or delete.

## Delete unused AI models

AI models are usually the largest consumers of disk space.

1. In **Files**, open the **Home** directory.
2. Delete unused models from the **Huggingface** and **Ollama** folders. These folders are named similarly to their corresponding apps for easier identification.

## Remove unused application images

Application and package images accumulate on the system over time. On Olares 1.12.6 and later, you can list unused images with `olares-cli` and remove them:

```bash
olares-cli doctor images --unused --no-headers | awk '{print $1}' | xargs -r sudo crictl rmi
```

:::warning
Do not use `sudo crictl rmi --prune`. It removes images for all stopped apps, which may slow down app resume and cause UI display issues.
:::

## Verify recovered space

Return to Dashboard and confirm that disk usage has dropped. If the system still reports low disk space, repeat the steps above or expand the system storage. For more information, see [Expand Olares system storage](best-practices/expand-storage-in-olares.md).

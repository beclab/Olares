---
outline: [2, 3]
description: Find answers to common questions about using Olares and community apps.
---
# Usage FAQs

Find answers to common questions about daily usage, applications, and system management.

## Applications

### What apps can I run in Olares?
The [Olares Market](https://market.olares.com/) maintains popular open-source apps like Ollama, ComfyUI, and Open WebUI.

If you have Docker experience, you can manually [deploy apps](../../developer/develop/tutorial/index.md) not listed in the Olares Market in a testing environment.

### Can I play games on my Olares device?
Yes. The Steam Headless app transforms your Olares device into a gaming server. It uses Proton to run the majority of the Steam game library on Linux, including most top Windows titles.

* [**Streaming:**](../../use-cases/stream-game.md) You can run games locally on Olares and stream them to devices like phones and tablets.
* [**Direct play:**](../../use-cases/play-games-directly.md) You can connect a monitor, keyboard, and mouse directly to the Olares device to play games without streaming.

### How do I access the Windows environment in Olares?
You can run a Windows VM from the Olares Market and access it using any standard RDP client.

For detailed instructions, refer to [Run a Windows VM on your Olares device](../../use-cases/windows.md).

### Can I develop apps on Olares?
Yes. You can install [Studio](../../developer/develop/tutorial/index.md) to code directly in your browser or connect your local VS Code to the device. This provides a development experience similar to your local machine but leverages the greater power of your server hardware.

## Storage
### If I add new disks to a running Olares machine, will Olares use them automatically?
It depends on the type of drive:
* **USB drives:** Yes, these are automatically mounted and will appear immediately in the Files app.
* **Internal drives:** No, internal HDDs or SSDs are not automatically added to the storage pool. You must configure them manually.
* **SMB shares:** Network storage can be added via **External** > **Connect to server** in Files.

For detailed instructions on adding internal drives, see [Expand storage in Olares](../best-practices/expand-storage-in-olares.md).

## Multi-node clusters

### How do I add more machines to my cluster?
By default, Olares installs as a single-node cluster. However, you can install Olares as a master node and then add worker nodes to create a scalable, multi-node cluster.

Note that this is currently an Alpha feature and works on Linux only. For detailed steps, please refer to [Install a multi-node Olares cluster](../best-practices/install-olares-multi-node.md).



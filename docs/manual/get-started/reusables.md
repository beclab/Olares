---
search: false
---

```bash
curl -fsSL https://olares.sh |  bash -
```

:::tip Root user password
During the installation, you may be prompted to enter your root password.
:::

:::info Errors during installation?
If an error occurs during installation, use the following command to uninstall first:

```bash
olares-cli uninstall --all
```

After uninstalling, retry the installation by running the original installation command.
:::

## Prepare Wizard URL

At the end of the installation process, you will be prompted to enter your domain name and Olares ID.

![Enter domain name and Olares ID](/images/manual/get-started/enter-olares-id.png)

For example, if your full Olares ID is `alice123@olares.com`:

- **Domain name**: Press `Enter` to use the default domain name or type `olares.com`.
- **Olares ID**: Enter the prefix of your Olares ID. In this example, enter `alice123`.

Upon completion of the installation, the initial system information, including the Wizard URL and the initial login password, will appear on the screen. You will need them later in the activation stage.

![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

## Next step: Protect your Olares ID

You're almost ready to start using Olares! Before diving in, it's crucial to ensure your Olares ID is securely backed up. Without this step, you won't be able to recover Olares ID if needed.

- [Back up your mnemonic phrase](/manual/larepass/back-up-mnemonics.md)

:::info Having trouble with installation?  
If you encounter issues during the installation process, feel free to [submit a GitHub Issue](https://github.com/beclab/Olares/issues/new). Please include the following information when submitting:

- The platform or environment you're using (e.g., Ubuntu, Docker, WSL, etc.).
- The installation method (script installation or Docker image).
- Detailed error information (including logs, error messages, or screenshots).  
  :::

## System requirements

Make sure your device meets the following requirements.

- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 150GB of available SSD storage. (The installation will fail if an HDD (mechanical hard drive) is used instead of an SSD.)
- Supported systems:
  - Ubuntu 22.04-25.04 LTS
  - Debian 12 or 13

:::info Version compatibility
While these specific versions are confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::
<!--Sync files, reused in LarePass > Manage files-->
## Sync files to local computer

With LarePass desktop, you can sync cloud files (organized by libraries or folders) to your local computer. This creates a corresponding folder on your machine. After set up, your files will stay updated bi-directionally in real time.

:::tip Note
The **Sync to local** feature is only available for libraries or folders within the **Sync** directory.
:::

### Create a library

Library is the fundamental unit for organizing, syncing, and sharing your digital content. Each user is automatically provided with their own personal library (My Library) as a starting point. 

To create a new library:

1. To the right of **Sync**, click <i class="material-symbols-outlined">add_circle</i> to open the **New library** dialog.
2. Enter a name for the library and click **Create**.

### Enable synchronization

To enable sync for a library or folder: 

1. Open LarePass desktop and locate the **Sync** directory.
2. Hover your mouse over the target library or folder, click <i class="material-symbols-outlined">more_horiz</i>  that appears on the right, and then click **Sync to local**. 

    ![Sync files to local](/images/manual/olares/sync-files-local.png#bordered){width=58%}
    
3. In the **Sync library** popup window, set the file download location, and then click **Confirm**. 
    
Syncing will begin immediately. Once completed, a green checkmark will appear on the bottom-left corner of the folder icon, indicating that the sync is finished.

### Manage synchronization

After setting up synchronization, you can manage your files and control the sync status with the following operations:

- If you want to quickly locate the sync directory on your local drive, hover your mouse over the target library or folder, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Open local sync folder**. The system will directly open the folder's location on your computer.

- If you no longer need to sync a folder, hover your mouse over it, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Unsychronize**.

- If you want to temporarily stop data transfer, click <i class="material-symbols-outlined">pause_circle</i> to the right of the **Sync** directory. All sync tasks will be paused.
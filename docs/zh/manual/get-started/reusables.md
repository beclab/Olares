---
search: false
---

```bash
curl -fsSL https://cn.olares.sh |  bash -
```

:::tip root 用户密码
安装过程中，可能需要输入 root 用户密码。
:::

:::info 安装遇到报错？
如果安装过程中出现错误，请先执行以下命令卸载：

```bash
olares-cli uninstall --all

```

卸载完成后，重新运行安装命令进行安装。
:::

## 配置 Wizard

在安装 Olares 的核心服务之前，需要输入在 LarePass 中注册的 Olares ID 前缀。如果你的 Olares ID 为 `alice123@olares.cn`，输入 `alice123` 即可。

![输入 Olares ID](/images/zh/manual/get-started/enter-olares-id.png)

安装完成后，屏幕将显示初始系统信息，包括向导地址和初始一次性密码。这些信息在后续激活步骤中会用到。

![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

## 下一步：安全保存 Olares ID

你已经准备好开始使用 Olares！在此之前，请务必确保 Olares ID 已安全备份。如果不备份，你将无法在需要时恢复 Olares ID。

- [备份助记词](../larepass/back-up-mnemonics.md)

## 系统要求

请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：不少于 150GB 的可用磁盘空间，需要使用SSD硬盘安装，使用HDD（机械硬盘）将会导致安装失败
- 支持的系统版本：
    - Ubuntu 22.04-25.04 LTS
    - Debian 12, 13

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台上安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 同步文件至本地

通过LarePass桌面端，你可以将云端的文件（以资料库或文件夹为单位）直接同步到本地电脑。系统将在你的本地电脑上创建一个对应的同步文件夹。设置完成后，文件将自动实时保持双向同步。

:::tip 注意
**同步至本地** 功能仅适用于 **同步** 目录下的资料库或文件夹。
:::


### 开启同步

若要为一个资料库或文件夹开启同步，完成以下步骤：

1. 打开LarePass桌面端，找到**同步**目录。
2. 将鼠标悬停在您想要同步的资料库或文件夹上，点击右侧出现的 <i class="material-symbols-outlined">more_horiz</i>，点击**同步至本地**。

   ![同步文件至本地](/images/manual/olares/sync-files-local-1.png#bordered)

3. 在**同步库**弹窗中，设置文件的下载位置，然后点击**确认**，同步随即开始。
    
完成后，该文件夹图标的左下角会出现一个绿色对勾，表示同步已完成。

### 管理同步

设置同步后，你可以通过以下操作来管理文件和控制同步状态：

- 如果您想快速定位到本地硬盘上的同步目录，将鼠标悬停在您想要定位的资料库或文件夹上，点击右侧出现的 <i class="material-symbols-outlined">more_horiz</i>，点击**打开本地同步文件夹**。系统将直接打开该文件夹所在的本地路径。

- 如果你不再需要同步某个文件夹，将鼠标悬停在该文件夹上，点击右侧出现的 <i class="material-symbols-outlined">more_horiz</i>，点击**取消同步**。

- 如果你需要暂时停止数据传输，点击**同步**目录右侧的 <i class="material-symbols-outlined">pause_circle</i>，即可一键暂停所有同步任务。
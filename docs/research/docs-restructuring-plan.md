# Olares OS 文档重组实施计划

> 依据：飞书《Olares OS documentation restructuring_Content map》表格（wiki 节点 F49owc7wQiBIdCkwtUicqsT2nxb，sheet「Content map」，172 行）
> 目标：按 Content map 的新 IA 重组文档站，**尽量只改导航**，最小化路由变动；允许新增 md 文档，但新文件必须按「主题/任务」命名路径，不按 app 名称存储内容。

---

## 1. 核心策略（三条原则）

0. **新增内容页一律扁平放在 `manual/` 根下**（用户确认，2026-09-08）：与 overview.md、glossary.md 同级，按任务/主题命名 slug。导航层级调整不需要移动文件、不产生 301。现有文件仍留在原路径不动。Phase 3 新页路径示例：`manual/how-access-works.md`、`manual/use-usb-drive.md`、`manual/mount-local-disk.md`、`manual/migrate-shared-apps.md`、`manual/password-and-devices.md`。
1. **所有现有 .md 文件原地不动**。新 IA 只体现在 `en.ts` / `zh.ts` 的 sidebar 里，sidebar 分组与磁盘目录解耦，分组名不需要对应目录名。
2. **能复用现有页面做 landing 的，不新建 stub 页面**；纯分组节点用无链接的 collapsible group。Content map 中「需要新增」的 Navigation 行，绝大多数只是分组。
3. **内容合并时，选信息量最大的现有页面作为 canonical，其余路径 301 重定向到它**；只有当现有路径语义严重不符时才新建 canonical 页面。

## 2. 路由保护机制（现状）

- 唯一信源：`docs/.vitepress/theme/redirects.ts`（`redirects` + `temporaryRedirects`）。
- 消费方：客户端重定向（`theme/index.ts`）+ `sync-redirects.mjs` 自动生成 `_redirects.nginx`（release 构建）和 `vercel.json`（Vercel 预览/生产）。`predev` / `prebuild` 会自动跑 sync，本地只需改 `redirects.ts`。
- 重定向是**路径级**的，不支持 anchor。涉及「从某页 #anchor 移出内容」的场景，原页面保留一个指向新页的链接，不做重定向。

## 3. Phase 1：仅导航重组（对齐 P0，零 URL 变化）

只改 `docs/.vitepress/en.ts` 和 `zh.ts` 的 `/manual/` sidebar（zh 同步）。新 L1 顺序按 Content map：Overview → FAQs → Get started → Tutorials → Accounts and access → Apps → Files and data → Personalize Olares → System → Help and troubleshooting → Glossary。

### 3.1 目标 sidebar 与现有页面映射

| 导航节点 | 链接（现有路径，不变） | 说明 |
|---|---|---|
| **Overview** | `/manual/overview` | L1 重命名（原 What is Olares），链接不变 |
| ├ Update notes ▸ Olares 1.12.6 | `/manual/update-guides/1.12.6` | 不变 |
| ├ What's new in docs | `/manual/release-notes` | 不变 |
| └ ~~Get support~~ | — | 无页面（未来 ticket 内容），本期不加 |
| **FAQs** | `/manual/help/faqs` | 从 Overview 下移出升为 L1 |
| ├ About Olares | `/manual/help/olares` | 不变 |
| ├ Installation and activation | `/manual/help/installation` | 不变（P1 拆分见 Phase 3） |
| └ Using Olares | `/manual/help/usage` | 不变（P1 补充见 Phase 3） |
| **Get started** | `/manual/get-started/` | 不变 |
| ├ Create an Olares ID | `/manual/get-started/create-olares-id` | 不变（P0 合并见 Phase 2-1） |
| ├ Install Olares ▸ | `/manual/get-started/install-olares` | 平台子项保持现状；macOS/Windows/PVE/LXC/RPi 继续隐藏（map 标注「隐藏」） |
| ├ Join an existing Olares as a member | `/manual/get-started/join-olares` | 不变 |
| └ What's next | `/manual/get-started/next-steps` | 不变（P0 微调） |
| **Tutorials** | `/manual/best-practices/` | 复用现有 index.md 做 landing，不新建 |
| ├ Activate a device using Olares CLI | `/manual/best-practices/activate-olares-using-cli` | 不变 |
| ├ Install a multi-node Olares cluster | `/manual/best-practices/install-olares-multi-node` | 不变 |
| ├ Install Olares on PVE with GPU passthrough | `/manual/best-practices/install-olares-gpu-passthrough` | 继续隐藏 |
| └ Install a specific CUDA version | `/manual/best-practices/install-specific-cuda-version` | 不变 |
| **Accounts and access** | 无链接分组 | 纯分组 |
| ├ **Olares ID**（分组） | | |
| │ ├ Manage Olares IDs in LarePass | `/manual/larepass/manage-accounts` | 不变 |
| │ ├ Set up a custom-domain Olares ID | `/manual/best-practices/set-custom-domain` | 不变（P0 合并见 Phase 2-8，吸收 Space 域名页） |
| │ ├ Manage your Olares Space account and billing | `/manual/space/manage-accounts` | P0 合并 canonical（Phase 2-13，吸收 billing） |
| │ └ Back up your mnemonic phrase | `/manual/larepass/back-up-mnemonics` | 不变 |
| ├ **Access Olares**（分组） | | |
| │ ├ How access to Olares works | 待新增（Phase 3-1） | P1 |
| │ ├ Access on your local network | `/manual/best-practices/local-access` | P0 合并 canonical（Phase 2-2） |
| │ ├ Access remotely with LarePass VPN | `/manual/larepass/private-network` | P0 合并 canonical（Phase 2-3） |
| │ ├ Connect to Olares via SSH | `/developer/reference/access-olares-terminal` | 跨站链到 developer 页，**不改路由** |
| │ ├ Create and manage team members | `/manual/olares/settings/manage-team` | 页内链接到 Roles and permissions |
| │ └ Change password and manage signed-in devices | `/manual/olares/settings/my-olares` | P1 拆分（Phase 3-6） |
| └ **Passwords**（分组） | | |
| │ ├ Create and manage Vault items | `/manual/olares/vault/vault-items` | 不变 |
| │ ├ Share Vault items securely | `/manual/olares/vault/share-vault-items` | 不变 |
| │ ├ Autofill passwords with LarePass | `/manual/larepass/autofill` | 不变 |
| │ └ Generate two-factor authentication codes | `/manual/larepass/two-factor-verification` | 不变 |
| **Apps** | `/manual/olares/` | 复用现有 index.md |
| ├ Install, update, and remove apps | `/manual/olares/market/market` | P0 合并 canonical（Phase 2-4） |
| ├ Clone an app | `/manual/olares/market/clone-apps` | 不变 |
| ├ Configure environment variables | `/manual/olares/settings/manage-app-env` | P0 合并 canonical（Phase 2-5） |
| ├ Allocate accelerator resources | `/manual/olares/settings/gpu-resource` | 不变（微调） |
| ├ Change an app's web address | `/manual/olares/settings/custom-app-domain` | 不变 |
| ├ Make an app available on the local network | `/manual/olares/settings/overlay-gateway` | P1 重写（Phase 3-7） |
| ├ Control who can access an app | `/manual/olares/settings/manage-entrance` | 不变（微调） |
| ├ Connect an AI app to a model service | `/manual/best-practices/connect-ai-apps` | 不变（微调） |
| ├ About shared applications | `/manual/olares/market/shared-apps` | P1 拆分（Phase 3-4） |
| └ Migrate legacy shared applications | 待新增（Phase 3-5） | P1 |
| **Files and data** | `/manual/olares/files/` | 复用现有 index.md |
| ├ **Work with files**（分组） | | |
| │ ├ Upload, edit, and download files | `/manual/olares/files/add-edit-download` | P0 合并 canonical（Phase 2-6） |
| │ ├ Compress and extract files | `/manual/olares/files/compress-extract-files` | 不变 |
| │ ├ Share files | `/manual/olares/files/share-files` | 不变（微调） |
| │ ├ Sync files with a computer | `/manual/olares/files/sync-files` | 不变 |
| │ ├ Configure file search | `/manual/olares/settings/search` | 不变 |
| │ └ Manage shared AI model files | `/manual/olares/files/files-common` | 不变（微调） |
| └ **Connect external storage**（分组） | | |
| │ ├ Connect an SMB share | `/manual/olares/files/mount-SMB` | P0 合并（anchor 级，Phase 2-7） |
| │ ├ Connect an NFS share | `/manual/olares/files/mount-nfs` | 不变 |
| │ ├ Connect cloud storage | `/manual/olares/files/mount-cloud-storage` | P0 合并（anchor 级，Phase 2-7） |
| │ ├ Use a USB drive | 待新增（Phase 3-2） | P1 |
| │ └ Mount a local disk | 待新增（Phase 3-3） | P1 |
| **Personalize Olares**（分组，P2 整组内容微调，本期只摆导航） | | |
| ├ Change language and appearance | `/manual/olares/settings/language-appearance` | P2 重写 |
| ├ Use the Olares Desktop | `/manual/olares/desktop` | P2 合并 |
| └ Create an Olares profile | `/manual/olares/profile` | P2 重写 |
| **System**（分组） | | |
| ├ Check system and app resource usage | `/manual/olares/resources-usage` | 不变 |
| ├ View Olares Space usage and system status | `/manual/space/manage-olares` | P1 合并（anchor 级） |
| ├ Update Olares | `/manual/larepass/manage-olares#upgrade-olares` | P0：以 LarePass 步骤为 canonical |
| ├ Reactivate Olares | `/manual/larepass/activate-olares` | 页面已存在，仅换导航位置 |
| ├ Manage your Olares device | `/manual/larepass/manage-olares` | P0 合并 canonical（anchor 级） |
| ├ Back up and restore Olares | `/manual/olares/settings/backup` | P0 合并 canonical（Phase 2-9） |
| └ **Storage**（分组） | | |
| │ ├ Free up disk space | `/manual/help/ts-free-disk-space` | P0 合并 canonical（见 4.3 决策） |
| │ └ Expand Olares system storage | `/manual/best-practices/expand-storage-in-olares` | P1 拆分后收窄范围 |
| └ **Advanced administration** | `/manual/olares/controlhub/` | 复用现有 index.md |
| │ ├ Manage workloads | `/manual/olares/controlhub/manage-workload` | 不变 |
| │ ├ Inspect and manage containers | `/manual/olares/controlhub/manage-container` | P1 微调 |
| │ ├ Manage resource configurations | `/manual/olares/controlhub/manage-resource` | P2 合并 manage-middleware |
| │ ├ Configure host name resolution | `/manual/olares/settings/set-up-hosts` | P1 微调 |
| │ ├ Manage repository mirrors | `/manual/olares/settings/developer` | P2 拆分后改链新页 |
| │ ├ Set system environment variables | `/manual/olares/settings/developer` | P1 拆分后改链新页 |
| │ ├ Change the reverse proxy | `/manual/olares/settings/change-frp` | 不变 |
| │ ├ Configure video playback | `/manual/olares/settings/video` | P2 微调 |
| │ └ Run commands in the Control Hub terminal | `/manual/olares/controlhub/terminal` | 不变 |
| **Help and troubleshooting** | `/manual/help/` | 复用现有 index.md |
| ├ Collect diagnostic information | `/manual/help/request-technical-support` | P0 合并（anchor 级，Phase 2-10） |
| ├ **Troubleshooting** | `/manual/help/troubleshooting-guide` | P0 重写为分类 landing（页内分类，不进导航） |
| └ Known issues | `/manual/help/known-issues` | 不变 |
| **Glossary** | `/manual/glossary` | 不变 |

### 3.2 Phase 1 同时要做的小内容修改（P0 微调，非合并）

- `get-started/next-steps.md`：补全「去哪做应用/文件/本地 AI/备份/团队邀请」的路由指引。
- `olares/settings/manage-team.md`：页内加 Roles and permissions 链接。
- `help/troubleshooting-guide.md`：重写成分类 landing（Apps and Market / AI and model runtime / Network access and domains / Storage backup and files / Accounts and authentication / Olares One hardware），分类下放现有 ts-* 页链接。
- `market/market.md` 等所有被 nav 换位置的页面：页内 frontmatter/上下游链接不动（URL 未变）。

## 4. Phase 2：P0 内容合并（新增 redirects.ts 条目）

每个合并：把被吸收页面的内容并入 canonical 页 → 删除被吸收页（EN + ZH 同步）→ 在 `redirects.ts` 加 EN 和 `/zh/` 两条。

> **Batch 1 进度（2026-09-08）**：2-1、2-5 已完成（build 通过）；2-4 已取消（隐藏页不处理）。

| # | canonical（保留） | 被吸收（删除+重定向） | 备注 |
|---|---|---|---|
| ~~2-1~~ ✅ | `manual/get-started/create-olares-id.md` | `manual/larepass/create-account.md` | 首次创建身份的单一主题；账户日常管理留在 manage-accounts |
| 2-2 | `manual/best-practices/local-access.md` | `manual/get-started/local-access.md` | 合并 LarePass VPN/.local/本地 DNS/逐设备 hosts；Get started 的「Access Olares securely」入口移除，改由 Access Olares 分组承载 |
| 2-3 | `manual/larepass/private-network.md` | `manual/olares/settings/remote-access.md` | 用户连接步骤 + 管理员 VPN 策略分角色写在一页 |
| 2-4 ⏭️ | `manual/olares/market/market.md` | `manual/olares/market/purchase-paid-apps.md` | ~~已取消~~：该页不在导航中（隐藏页），用户确认不处理 |
| ~~2-5~~ ✅ | `manual/olares/settings/manage-app-env.md` | `manual/olares/controlhub/configure-env-var.md` | 安装时设置/安装后 Settings/Control Hub 高级路径合一 |
| 2-6 | `manual/olares/files/add-edit-download.md` | `manual/larepass/manage-files.md` | web + LarePass 客户端双路径，页内分区 |
| 2-8 | `manual/best-practices/set-custom-domain.md` | `manual/larepass/create-org-account.md`、`manual/space/host-domain.md`、`manual/space/manage-domain.md` | 端到端：域名设置→身份创建→域名成员。三个被吸收页面均加 301；Space 侧的「Add custom domain」导航组随区块解散移除 |
| 2-9 ✅ | `manual/olares/settings/backup.md` | `manual/olares/settings/restore.md`、`manual/space/backup-restore.md` | backup.md 已扩写为「Back up and restore data in Olares」：备份任务 + 本地/Olares Space/AWS S3/Tencent COS 恢复 + 恢复任务管理。restore.md 与 space/backup-restore.md EN/ZH 均已删除，301 → settings/backup；旧 `/space/backup-restore` 重定向直改最终目标避免链式。入站修复：developer/concepts/data.md（EN/ZH）、space/index.md（临时指向，2-13 解散时处理）、backup.md 内 Settings › Integrations 失效引用 → 改指 mount-cloud-storage.md。backup 页内 billing 链接层级已修。2026-09-08 build 通过 |
| 2-13 | `manual/space/manage-accounts.md` | `manual/space/billing.md` | 扩写为「Manage your Olares Space account and billing」；billing 301 过来。**Olares Space sidebar 区块解散**，`space/index.md` 301 到本页 |
| 2-7* ✅ | `mount-cloud-storage.md` = 云存储唯一任务页；`mount-SMB.md` = SMB 挂载任务页（含“Save SMB accounts”一节） | **两个 integrations 页全部 dissolve**（用户评审：不要为保页面而切分职责，能接受砍页面）。内容归位：云存储→mount-cloud-storage；SMB 账号→mount-SMB；Cookie（查看/删除）→wise/manage-cookies.md（该页已有完整流程，零新增）；Space 登录→space/manage-accounts、授权→space/manage-olares Before you begin、绑定→nft-image 自带。nav 两项已删（Connect cloud drives in LarePass / 管理存储集成），301：larepass/integrations 与 settings/integrations 均指向 mount-cloud-storage，nft-image 前置链接已改为自包含 | 2026-09-08 build 通过。2-13 Space 部分前置完成（连接/授权流程已在 space/ 页中，无遗留） |
| 2-10* | `manual/help/request-technical-support.md` | `settings/developer.md#export-system-logs`、`controlhub/manage-container.md#logs` 段落 | 扩写诊断收集页（UI/CLI/容器日志/保存位置/安全分享） |
| 2-11* ✅ | `manual/larepass/manage-olares.md` | `settings/my-olares.md` 硬件/重启关机段落 | 重启/关机两入口已合并进 manage-olares「Restart or shut down Olares」（LarePass + Settings > My hardware 双入口，均需在 LarePass 确认）；ZH 同步。2-12 Space 指针段并入 space/manage-olares 既有指针，无新增。SSH 重置指向 one/access-terminal-ssh.md#reset-ssh-password。**my-olares.md EN/ZH 已删除**，301 → /manual/password-and-devices。硬件内容（硬件详情/工作模式/限制 CPU 频率/自动开机）→ 新建 `one/hardware-settings.md`（+zh，One 专属故放 one/ 而非 manual/ 根），nav 挂在 one 侧栏「管理 BIOS 和 EC」后；入站修复：resources-usage、comfyui（#limit-cpu-frequency）、release-notes。**Communication and feedback / Acknowledgements 两节无归并目标，直接放弃**（用户原则：能合并就合并，但这两节没有可并入的页面）。密码/网络策略/已登录设备/退出登录 → `manual/password-and-devices.md`（+zh），en/zh 侧栏原「Change your password and manage signed-in devices」项已改指新页。2026-09-08 build 通过 |
| 2-12* | `manual/space/manage-olares.md` | `settings/my-olares.md#olares-space` 段落 | 「View Olares Space usage and system status」 |

\* anchor 级合并不产生重定向条目。~~2-11/2-12 后 `my-olares.md` 剩余密码/设备/Space 小节由 Phase 3-6 再拆分。~~（2-11/2-12 已于 2026-09-08 执行完毕，my-olares.md 已整体处理）

### 4.3 已确认的 canonical 决策

- **Free up disk space**：`help/ts-free-disk-space.md` 实际不存在（2026-09-08 核实），故新建 canonical：`manual/free-up-disk-space.md`（扁平放 manual/ 根下，用户确认不放 help/ 也不按目录层级组织）。Storage 导航和 Troubleshooting「Disk space is full」都链到它；镜像清理步骤用 `olares-cli doctor images --unused` + `crictl rmi`，明确警告不要用 `crictl rmi --prune`；模型清理在 Files > Home 的 Huggingface/Ollama 文件夹。
- **Local access canonical**：`best-practices/local-access.md`，get-started 版本 301。
- **manage-olares 精简 + 更新 canonical**（用户 2026-09-08：「升级都可以删掉了吧」「reset ssh password 移到 Olares One」「Access Olares management 太冗长」）：manage-olares EN/ZH 已删 Upgrade 与 Reset SSH 两节、压缩 Access 节（去 system.png 截图和能力清单）。Update canonical = 原隐藏页 `settings/update.md` 转正（去 noindex，双入口 Tabs：LarePass / Settings，含 olaresd 手动升级），nav「Update Olares」已改指；one/update.md 为 One 教程不动。Reset SSH canonical = one/access-terminal-ssh.md#reset-ssh-password（含激活弹窗，经 reusables 覆盖），manage-olares 侧删除消除与 one/ 新流程的冲突描述。
- **settings/developer.md 拆分**（用户 2026-09-08：「这个也可以拆分哈」）：系统级环境变量 → 新建 `manual/manage-system-env.md`（+zh，扁平），nav 挂 Apps 组（manage-app-env 之后）；导出系统日志节删除（canonical = help/request-technical-support.md，同一 reusable，developer.md  intro 留指针）；developer.md 保留仓库+镜像管理，intro 留两个指针。
- **Olares Space 区块解散**：所有 Space 页面合并到 map 指定位置（2-8 域名页、2-9 备份还原、2-12 状态监控、2-13 账户与计费），`space/index.md` 301 到 `space/manage-accounts.md`（zh 同步）。

## 5. Phase 3：P1 拆分 / 新增 / 重写

新文件一律按主题/任务命名，不用 app 名：

| # | 新文件 | 来源 | 导航挂载点 |
|---|---|---|---|
| 3-1 | `manual/how-access-works.md` | 新增（解释 Olares 访问机制：本地/远程/VPN/SSH 的关系） | Access Olares |
| 3-2 ✅ | `manual/use-usb-drive.md` | 从 expand-storage-in-olares.md 拆出 | Connect external storage。已建（+zh），nav 置组首 |
| 3-3 ✅ | `manual/mount-local-disk.md` | 从 expand-storage-in-olares.md 拆出（挂载 HDD/SSD 到 /olares/share，含临时/持久/卸载） | Connect external storage。已建（+zh），nav 第二项。expand 页已收缩为纯 disk-extend 系统扩容（H1 改「Expand Olares system storage」，补交叉链接 mount-SMB/local-disk/USB，悬空引用「disk 命令文档」补为 /developer/install/cli/disk.md 真实链接）；SMB 节删除（mount-SMB.md 为 canonical）。用户提示触发：「这一页也可以拆分哈，前面有讲怎么连 SMB」 |
| 3-4 | 拆分 `manual/olares/market/shared-apps.md` | 架构历史移入 `update-guides/1.12.6.md`（或 release notes），正文保留「About shared applications」 | Apps |
| 3-5 | `manual/migrate-shared-apps.md` | 从 shared-apps.md#migrate 段扩写（按 app 的迁移路径+数据保护检查清单） | Apps（Tutorial 性质） |
| 3-6 ✅ | `manual/password-and-devices.md` | 从 my-olares.md 拆出 change-password / set-network-access-policy / devices / log-out 四节 | Access Olares。已建（+zh），含入口说明（Settings 左上角头像进入 My Olares）。**完成于 2-11 执行时** |
| 3-7 | 重写 `manual/olares/settings/overlay-gateway.md` | 强调「让应用可被局域网发现」，与「本地访问 Olares 本身」区分 | Apps |
| 3-8 | 新增 `manual/help/ts-login-errors.md` | 从 `help/installation.md#login-and-authentication-error-messages` 移出。**anchor 无法 301**，installation.md 原位置留一行链接 | Troubleshooting（Accounts and authentication 分类） |
| 3-9 | 补充 `manual/help/usage.md` | 新增常见使用问题（简短可答，不写成完整流程） | FAQs |
| 3-10 | 微调：`expand-storage-in-olares.md` 收窄为「Expand Olares system storage」（olares-cli disk extend + 数据破坏警告）；`manage-container.md` 诊断导出改为链到 Collect diagnostic information | — | — |

## 6. Phase 4：P2

- 重写 `language-appearance.md`（语言+主题+壁纸+Dock）；合并 `desktop.md` 与 next-steps 的桌面介绍；重写 `profile.md`。
- `manage-middleware.md` 并入 `manage-resource.md`，加 301。
- 从 `developer.md` 拆出 repository mirrors、system env vars 两个新页（`manual/olares/settings/repo-mirrors.md`、`system-env-vars.md`），developer.md 瘦身。
- 微调 `video.md`、`ts-cs-app-reappears.md`（保留版本徽章，注明受影响版本停止支持后归档）。

## 7. 范围外 / 待办

- **Olares Space 区块**（已确认解散）：全部页面按 Phase 2-8 / 2-9 / 2-12 / 2-13 合并。唯一遗留：`manual/space/create-olares.md` 目前无导航入口，文件暂留，是否 301 到 get-started 安装流程待确认。
- **Get support**：map 中为未来 in-product ticket 占位，本期不加入导航。
- **SSH 页路由**（2026-09-08 用户指示已执行）：页面从 `developer/reference/access-olares-terminal.md` 移到 `manual/access-olares-terminal.md`（扁平根下，git mv 保留历史），旧路径 301 到新路径（EN/ZH）。developer.zh.ts sidebar 中的活跃入口同步改链。
- `olares/wise/` 全套文档当前已在导航中注释隐藏，本次不动。
- Olares One（`one.en.ts` / `one.zh.ts`）与 Use cases、Developer guide 导航均不动。

## 8. 执行顺序与验证

1. Phase 1 先落地（纯导航），单独 PR，可独立验证路由零变化。
2. Phase 2 每个合并单独 commit，redirects.ts 成对（EN/ZH）追加。
3. 验证：
   - `cd docs && npm run dev`：sidebar 结构、每个新链接可点、隐藏页仍隐藏。
   - `npm run build` 通过；检查 sync-redirects 生成的 `vercel.json`/`_redirects.nginx` 只增不改。
   - 逐条 curl 验证 301：旧路径 → canonical（本地 preview 验证客户端重定向，Vercel 预览验证 308）。
   - 全文搜索被吸收文件的旧路径引用（`grep -r` 旧 slug），更新页内交叉链接。
   - EN/ZH 同步检查：两侧文件数、导航条目数一致。

# Goose 验收记录 — 2026-10-13

## 结论

已覆盖的用户使用流程通过，包括用户多次确认正常的网页“停止 → 恢复”。由于正式市场新装及部分可选环境尚未覆盖，仍不能宣布全部发布场景完成验收。独立 CLI `market restart` 接口超时属于工具／API 诊断，不算 Goose 用户流程失败；旧部署时区差异也按用户要求排除。

测试环境为现有部署：Olares `1.12.7-20260908`、Goose 服务端／Desktop `1.54.0`、Terminal `0.0.14`、amd64。AI Router 默认模型为 `unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL`，推理引擎并发数为 1。正式 Market 暂无该应用，未验证从正式市场全新安装。本轮不覆盖 arm64，以及 macOS、Linux Desktop。

结果区分自动执行与人工验证。用户已明确确认：Desktop 重连、错误连接密钥拒绝与恢复、VPN 中断与恢复三项均通过。报告不包含任何密钥。

汇总：16 条通过、5 条部分验证、0 条失败、1 条阻塞。旧部署时区差异不计为版本失败。

## 测试用例与结果

| ID | 优先级 | 前提条件 | 操作步骤 | 预期结果 | 实际结果 | 状态 |
|:---|:---|:---|:---|:---|:---|:---|
| GOOSE-001 | P0 | 尚未安装 Goose；Router 前提满足 | 在 Market 搜索 Goose，打开应用页面，点击获取并安装，等待完成。 | 安装完成并进入运行中状态。可从 Launchpad 打开 Goose Terminal；设置中可找到 Goose ACP，且不将其描述为浏览器聊天界面。 | 正式 Market 目录中没有 Goose。未执行正式市场全新安装，也未替换现有安装。 | 阻塞 |
| GOOSE-002 | P0 | 全新安装；未修改服务商配置 | 查看应用环境变量，在 Goose Terminal 启动 `goose session`，发送简短问候。 | `GOOSE_PROVIDER=custom_olares_router`、`GOOSE_MODEL=default-chat`。无需填写默认连接的模型 API 密钥即可获得完整回复；Router 证据确认使用其默认模型。 | 现有应用使用 `custom_olares_router` 和 `default-chat`；Router 与 Goose 均成功返回就绪标记，无需新填模型 API 密钥。未验证全新安装的默认值。 | 部分验证 |
| GOOSE-003 | P0 | Desktop 已安装；VPN 已开启；Goose 运行中 | 从设置复制 ACP HTTPS 地址，从 Terminal 获取连接密钥；配置外部服务端、地址、密钥和 `/config/workspace`，新建对话发送就绪确认消息。 | 客户端成功连接并收到完整回复。当前会话使用远程服务，无需为既有 Router 配置重新执行服务商首次配置。 | Windows Desktop 在新远程会话中完成对话；用户另行确认默认 Router 重连成功。 | 通过 |
| GOOSE-004 | P0 | Desktop 已成功连接 | 将客户端密钥替换为错误值，应用配置并建立新连接；再恢复正确值重新连接。 | 错误密钥被拒绝，不能进行远程对话或文件操作；恢复正确密钥后可正常使用。 | 错误密钥探测返回 401，正确密钥返回 406，表示认证通过但探测缺少 ACP 流式请求头。用户确认 Desktop 错误密钥拒绝及恢复流程通过。 | 通过 |
| GOOSE-005 | P0 | Desktop 已连接；模型支持工具调用；示例文件不存在或已隔离 | 执行指南中的项目简报提示词，批准文件操作；在 Terminal 中用 `ls` 和 `cat` 检查两份文件。 | 两份文件均在指定目录中实际存在。简报保留所有事实，没有编造负责人或时间安排。指定示例目录外无写入。 | Desktop 在 Olares 上创建两份英文示例文件。独立读取确认目标、两项交付物、日期和待确认负责人问题均保留，未编造负责人或时间安排。 | 通过 |
| GOOSE-006 | P1 | GOOSE-005 通过 | 发送追加“后续行动”的请求，再次读取文件。 | 磁盘文件实际更新，原事实保留，建议明确标注为提议。 | Desktop 追加三条明确标为提议的建议；原有事实完整保留，磁盘文件实际更新。 | 通过 |
| GOOSE-007 | P0 | Goose 运行中；Desktop 无进行中的请求 | 打开 Terminal，执行 `cd /config/workspace` 和 `goose session`，先发送问候，再请求读取简报，最后按 Ctrl+C 退出。 | 两次请求均使用默认服务商完成，正确读取文件；退出 CLI 不停止服务，Desktop 仍可连接。 | CLI 对话完成并退出。Terminal 页面／WebSocket 返回 200／101；实际 Terminal 协议收到 `BROWSER_TERMINAL_READY`，退出后 shell 返回 `TERMINAL_SHELL_ALIVE` 和 `TERMINAL_EXIT_RECOVERED`。TUI 提示 Ctrl+C 可能先中断请求、再按一次才退出，文档已修正。本项验证浏览器后端协议，未执行可视化浏览器点击流程。 | 通过 |
| GOOSE-008 | P1 | 简报存在；模型无进行中的请求 | 执行指南中的一次性 `goose run -t` 命令，比较执行前后的文件。 | 正确概括待确认事项，命令成功结束，文件未修改。 | 一次性命令读取简报并概括待确认事项。前后 SHA-256 一致：`13fde6b50ff750c3e91fe1cc3f541d7753d29eca7ae3e42e45aad5c0d45d625a`。 | 通过 |
| GOOSE-009 | P1 | 至少一个已完成的 Terminal 会话 | 执行 `goose session list`，选择实际 ID，运行 `goose session --resume --session-id ID`，询问先前对话内容。 | 恢复指定会话及其历史，不创建无关的空白会话。 | 恢复会话 `20261013_19` 后正确回忆 `CLI_SESSION_READY`；应用恢复启动后仍能恢复并回忆。 | 通过 |
| GOOSE-010 | P0 | 全新安装；两个目录开关均为 false；Home 和 External 有已知标记文件 | 请求读取 `/config/Home` 和 `/config/External` 中的标记文件，并从 Terminal 检查。 | 两个用户目录未通过这些路径挂载，Goose 无法读取标记文件，也不声称读取成功；持久化工作目录正常可用。 | 两个开关关闭时，Home／External 专用标记文件无法读取；Desktop 明确报告文件不可访问，没有编造内容。 | 通过 |
| GOOSE-011 | P0 | Home 中有测试副本；External 访问关闭 | 仅设置 `ALLOW_HOME_DIR_ACCESS=true`，保存并等待运行中，重新连接后读取 Home 标记文件。 | `/config/Home` 可访问，内容与标记一致；External 仍不可访问；智能体仅读取指定测试文件。 | 仅开启 Home 时，Goose 准确读取 `HOME_MARKER_GOOSE_ACCEPTANCE_20261013`；External 标记仍不可访问。 | 通过 |
| GOOSE-012 | P0 | External 中有测试副本；Home 访问关闭 | 仅设置 `ALLOW_EXTERNAL_DIR_ACCESS=true`，保存并等待运行中，重新连接后读取 External 标记文件。 | `/config/External` 可访问，内容与标记一致；Home 仍不可访问。 | 仅开启 External 时，Goose 准确读取 `EXTERNAL_MARKER_GOOSE_ACCEPTANCE_20261013`；Home 标记仍不可访问。 | 通过 |
| GOOSE-013 | P0 | 某个目录访问用例已通过 | 将对应开关恢复为 false，保存并等待运行中，在新会话中读取原标记文件和工作目录简报。 | 用户目录不再通过挂载路径开放，原标记文件未变，工作目录简报仍可访问。 | 关闭两个开关后，两个标记均不能再通过挂载路径读取，工作目录简报未变。卸载挂载后可能保留空目录，本轮按标记文件是否可访问判断权限，而非按目录是否存在判断。 | 通过 |
| GOOSE-014 | P1 | 默认配置；Router 模型只允许一个活跃请求；其他调用方空闲 | 确认 `GOOSE_DISABLE_SESSION_NAMING=true`，在 Desktop 新会话中依次发送四条消息，每次等待回复完成，必要时核对 Router 请求证据。 | 四条消息均完成，没有由自动标题生成引起的冲突；Goose 不发送标题生成请求；默认 New Chat 标题属于预期。其他调用方造成的冲突单独排查。 | 同一新 Desktop 会话连续完成四条顺序消息，包含文件任务，没有 503。自动标题关闭，会话保持 New Chat；模型实际并发配置确认为 1。 | 通过 |
| GOOSE-015 | P2 | 已有保存的会话 | 手动重命名会话，再次打开或恢复。 | 自定义标题保留，会话内容完整；重命名不需要 AI 标题生成请求。 | 通过支持的 CLI 手动将测试会话重命名为 `Goose-Acceptance-Verified`，停止及恢复启动后名称保留。 | 通过 |
| GOOSE-016 | P1 | 有可用的 OpenAI 兼容服务商 | 按指南配置服务商、准确模型 ID、不带 `/v1` 的根地址、请求路径和 API 密钥，保存并等待运行中；分别在 Desktop 和 Terminal 新会话中顺序发送问候。 | 两者均使用指定服务商和模型；请求路径正确，没有重复的 `/v1`；截图和报告中不泄露密钥。 | 文档中的 OpenAI 兼容根地址／请求路径配置连接用户本地模型成功，返回 `DIRECT_CONFIG_ACCEPTANCE_OK`。未测试真实第三方账号／密钥及切换后的新 Desktop 会话。 | 部分验证 |
| GOOSE-017 | P1 | GOOSE-016 通过；Router 已就绪 | 恢复 `custom_olares_router` 和 `default-chat`，移除不再使用的第三方密钥，保存并等待运行中；在两种界面中开始新会话。 | 两者均通过 AI Router 完成请求，不依赖已经移除的第三方密钥。 | 恢复默认 Router 配置后，CLI 返回 `ROUTER_RESTORED_AFTER_RESTART`；用户确认 Desktop 重连通过。本轮未新增第三方密钥。 | 通过 |
| GOOSE-018 | P0 | 已保存会话，简报已验证 | 记录文件内容和会话 ID，在应用页面点击停止并等待已停止，再点击恢复并等待运行中。重新连接，检查文件并恢复会话。自动化依次执行 `olares-cli market stop goose --watch` 和 `olares-cli market resume goose --watch`，分别校验最终状态为 Stopped、Running，不仅检查退出码；不使用 `market restart` 替代。 | 两个生命周期操作均完成，文件内容和会话历史保留，已有客户端密钥仍可认证，无需重新安装。 | 用户多次验证网页“停止 → 恢复”正常。自动检查另行确认文件、会话、手动名称和连接认证在停止／恢复后保留。独立的 `market restart` 接口超时在下方作为工具诊断保留，不计入应用验收失败。 | 通过 |
| GOOSE-019 | P1 | TIMEZONE 保持系统提供的默认值 | 记录用户的 Olares 时区，在 Terminal 执行 `printenv TZ` 和 `date`；具备条件时使用另一用户时区重复验证。 | TZ 跟随 `OLARES_USER_TIMEZONE`，观察到的时间使用对应时区，不强制使用 Asia/Shanghai；记录平台配置传播延迟。 | 用户变量 `OLARES_USER_TIMEZONE` 为 UTC，运行中容器的 `TZ` 却为 Asia/Shanghai。拟发布 Chart 已包含时区继承，但当前部署未体现。按用户要求，该旧部署观察不计为版本失败；新装时区继承仍待验证。 | 部分验证 |
| GOOSE-020 | P1 | Desktop 已连接 | 断开 VPN，尝试新请求，然后恢复 VPN 并重新连接。 | 服务不可达时明确报错，不声称任务执行成功；VPN 恢复后连接恢复；重试写入前确认可能已完成的部分操作。 | 用户确认 Desktop 断开／恢复 VPN 行为符合预期。另行验证的隔离不可达端点失败，实际服务健康接口返回 200。 | 通过 |
| GOOSE-021 | P1 | 可在不影响其他用户的情况下制造 Router 故障 | 使默认模型暂不可用，发送一条消息；恢复模型后在新会话中重试。 | 显示模型错误，不编造成功结果；Router 就绪后回复恢复；保留证据区分模型错误与连接密钥错误。 | 仅为 Goose 配置不存在的模型后，Router 明确返回 404；恢复 `default-chat` 后对话恢复。未停止共享模型，因此未复现真实模型不可用／503 的恢复流程。CLI 显示模型错误却返回退出码 0，自动化不能仅依赖退出码。 | 部分验证 |
| GOOSE-022 | P1 | 中英文文档及导航变更已准备好 | 分别按中英文指南从安装执行至文件验证，检查侧边栏、图库卡片、命令、设置和前提条件。 | 两种语言均可完成同一使用流程；导航进入正确语言页面；不存在失效链接、暴露的密钥或特定环境地址。 | 英文 Desktop 和中文 CLI 文件任务完成，事实完整。中文任务曾误选图片读取工具，随后改用 `cat` 并生成正确文件。已核对导航、语言链接、Markdown 渲染及设置标签；正式市场安装和已发布站点的导航尚未验证。 | 部分验证 |



## 待跟进问题

1. 部署拟发布版本的 Chart 后重新验证时区继承。当前 UTC／Asia/Shanghai 的差异已足以证明此部署的问题，因此未修改用户时区。
2. 将独立 CLI 重启接口超时作为工具／API 问题排查。用户流程验收和后续自动化均应使用“停止 → 恢复”，等待前一操作完成后再执行后一操作。
3. CLI 模型错误可能返回退出码 0，自动化需同时检查错误及结果内容。

## 最终状态

已恢复 `custom_olares_router`／`default-chat`，自动会话标题关闭，Home／External 访问关闭。ACP 入口仍为 `internal`，Terminal 入口仍为 `private`。测试文档和专用标记文件保留供检查，未修改用户时区或共享模型配置。

文档已按 v1.54.0 对齐 Desktop 字段：External Backend (ACP)、Use external backend、Backend Base URL 和 Remote Working Directory (optional)。整站构建仍受原有页面缺少图片影响，该问题与运行时验收分开记录。

## 独立 CLI／API 诊断：重启接口

本诊断不计入 Goose 应用验收失败。首次发现独立重启接口失败后，又通过相同 CLI／API 路径顺序补测三次。三次均请求超时，最终停在 stopped。第二次在请求超时时尚为 stopping，等待后确认变为 stopped，再执行恢复。第二、三次请求分别约 30.36 秒、30.34 秒超时。每次显式恢复启动都成功。

| 补测轮次 | 重启结果 | 最终状态 | 恢复结果 |
|:---|:---|:---|:---|
| 1 | Context deadline exceeded | Stopped，已无 Pod | 恢复启动成功 |
| 2 | Context deadline exceeded，约 30.36 秒 | Stopped | 恢复启动成功 |
| 3 | 请求超时，约 30.34 秒 | Stopped | 恢复启动成功 |

本次补测复现率为 3／3，可确认当前 CLI／API 路径稳定复现；尚不能据此认定 Olares 图形界面中的重启也存在相同问题。收尾检查健康接口为 200，两份示例文件校验值不变，手动重命名的会话保留，原连接密钥仍通过认证。

---
outline: [2, 3]
description: 根据现象查找应用、AI、网络、存储、账号和 Olares One 硬件对应的故障排查指南。
---

# 排查 Olares 故障

先找到发生故障的 Olares 功能分类，再根据看到的现象选择指南。你不需要预先判断问题是简单还是复杂。

修改系统前，请记录完整错误信息和发生时间，确认 Olares 与相关应用的版本，并查看[已知问题](./known-issues.md)。除非指南明确要求，否则不要重装 Olares 或删除数据。

## 应用与应用市场

应用不可见、无法安装或卸载、状态异常时，请查看此分类。

| 你看到的现象 | 排障指南 |
|---|---|
| 应用市场中缺少预期的应用 | [应用市场中缺少应用](./ts-missing-apps.md) |
| 应用独占模式下无法移除已暂停的应用 | [无法在应用独占模式移除已暂停应用](./ts-cs-app-reappears.md) |
| 应用市场和控制面板显示的应用状态不一致 | [控制面板启停应用后状态不一致](./ts-inconsistent-app-status.md) |
| 应用安装或更新失败、操作一直未结束，或操作后应用无法打开 | [应用安装、更新期间或之后失败](./ts-app-fails-after-update.md) |

## AI 与模型运行时

AI 应用或模型无法获得所需内存或显存资源时，请查看此分类。

| 你看到的现象 | 排障指南 |
|---|---|
| 可用内存不足，或停止应用后内存仍未释放 | [内存不足或没有释放](./ts-free-memory.md) |
| GPU 应用处于暂停状态，或因显存不足无法启动 | [GPU 应用安装或恢复后处于暂停状态](./ts-vram-shortage.md) |
| 模型一直未就绪，或推理引擎无法启动 | [模型或引擎未就绪](./ts-model-engine-not-ready.md) |

## 网络、访问与域名

无法从本地或远程访问 Olares、专用网络连接失败，或串流不稳定时，请查看此分类。

| 你看到的现象 | 排障指南 |
|---|---|
| LarePass 专用网络无法连接或无法访问 Olares | [LarePass 专用网络无法使用](./ts-larepass-vpn-not-working.md) |
| 激活期间或重启后无法访问 Olares | [网络尚未准备好或 Olares 连接错误](./ts-network-not-ready.md) |
| 修改自定义路由 ID 后，应用 URL 返回 404 | [自定义路由 ID 导致应用无法访问](./ts-custom-route-domain.md) |
| Steam 串流卡顿、延迟或不稳定 | [Steam 串流卡顿或延迟](./ts-steam-stream-lag.md) |

## 存储、备份与文件

系统磁盘空间不足，或应用与模型文件占用空间过多时，请查看此分类。

| 你看到的现象 | 排障指南 |
|---|---|
| 磁盘空间已满，或模型与镜像文件占用过多空间 | [清理磁盘空间](../free-up-disk-space.md) |

## 账号与身份验证

登录、密码、激活或身份验证失败时，请查看此分类。

| 你看到的现象 | 排障指南 |
|---|---|
| 忘记桌面登录密码 | [忘记桌面登录密码](./ts-forget-login-password.md) |
| LarePass 显示**系统错误** | [LarePass 显示“系统错误”](./ts-system-error.md) |
| 登录或激活出现明确错误信息 | [登录与激活错误信息](../login-and-activation-errors.md) |

## Olares One 硬件

Olares One 的 BIOS、启动或双系统出现问题时，请查看此分类。

| 你看到的现象 | 排障指南 |
|---|---|
| Olares One 状态灯不亮，或连接显示器后一直没有画面 | [Olares One 无法开机或没有画面](/zh/one/ts-no-power-or-display.md) |
| 尝试进入 BIOS 时屏幕保持黑屏 | [无法进入 BIOS（黑屏）](/zh/one/ts-bios-black-screen.md) |
| Windows 双系统无法启动或运行不符合预期 | [排查 Windows 双系统问题](/zh/one/dual-boot-windows-troubleshooting.md) |

## 没有找到对应问题

记录问题上下文，只收集相关日志。具体方法请参考[收集诊断信息](../collect-diagnostic-information.md)。

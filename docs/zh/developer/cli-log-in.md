---
outline: [2, 3]
description: 通过 olares-cli 登录 Olares 并管理 profile。涵盖交互式登录、查看和切换 profile、删除 profile，以及令牌的存储位置。
---

# 登录 Olares

使用 `olares-cli` 以 Olares 用户身份操作前，需要先登录并创建一个 profile。一个 profile 对应一个 Olares 实例和一个用户身份。登录后 CLI 会自动刷新令牌，只有刷新令牌失效时才需要重新登录。

本页介绍用户模式的登录。主机模式使用 root 和 kubeconfig，无需登录。

## 首次登录

1. 运行以下命令开始登录。把 `alice123@olares.com` 换成你自己的 Olares ID。

   ```bash
   olares-cli profile login --olares-id alice123@olares.com
   ```

2. CLI 提示 `password for <id>:` 时，输入你的 Olares 密码并按回车。输入内容不会显示。

3. 如果账户开启了两步验证，CLI 会再次提示 `two-factor code for <id>:`。输入 LarePass 中的 6 位验证码并按回车。

4. 确认 profile 已创建并处于登录状态。

   ```bash
   olares-cli profile list
   ```

   输出示例：

   ```text
      NAME                   OLARES-ID              STATUS
      laresprime@olares.com  laresprime@olares.com  logged-in
   *  alexmiles@olares.com   alexmiles@olares.com   logged-in
   ```

   开头的 `*` 标记当前 profile。

## 管理 profile

如果你要操作多个 Olares 实例或身份，每次登录都会新增一个 profile。用以下命令在它们之间切换。

| 任务 | 命令 |
|------|------|
| 列出所有 profile | `olares-cli profile list` |
| 查看当前身份 | `olares-cli profile whoami` |
| 切换到另一个 profile | `olares-cli profile use <name>` |
| 切回上一个 profile | `olares-cli profile use -` |
| 删除 profile 及其令牌 | `olares-cli profile remove <name>` |

## 令牌的存储位置

登录成功后，令牌会自动保存，无需手动管理。要清除令牌，请用 `olares-cli profile remove`，不要直接编辑文件。

| 操作系统 | 存储方式 |
|---------|---------|
| macOS | 钥匙串（Keychain） |
| Linux | `~/.local/share/olares-cli/` 下的 AES 加密文件 |
| Windows | DPAPI |

## 下一步

把 [Agent Skills](./cli-agent-skills.md) 安装到你的 Agent 中，即可用自然语言操作 Olares。

# Reusables（可复用片段）

本目录存放通过 `<!--@include: path/to/reusables/file.md#region-name-->` 在多个文档中引用的共享内容。

新增可复用片段时，请使用命名 VS Code region 包裹：

```markdown
<!-- #region region-name -->
可复用内容。
<!-- #endregion region-name -->
```

region 名称应稳定且具备描述性，避免源文件行号变化后影响引用。

- **local-domain.md**：局域网直连、hosts 映射、地址格式和故障排除。被 `manual/get-started/local-access.md`、`manual/best-practices/local-access.md` 和 `one/access-olares-via-local-domain.md` 引用。
- **larepass-vpn.md**：LarePass 专用网络步骤（下载、启用、确认连接类型）及常见问题（链接至故障排查文档）。被 `manual/get-started/local-access.md`、`manual/best-practices/local-access.md` 和 `one/access-olares-via-vpn.md` 引用。
- **sync-files.md**：同步文件至本地（引言与提示、创建库、开启同步、管理同步）。被 `manual/larepass/manage-files.md`、`manual/olares/files/sync-files.md` 引用。
- **export-system-logs.md**：通过设置 > 高级 > 日志导出系统日志的步骤。被 `manual/help/request-technical-support.md` 引用。
- **custom-domain.md**：自定义域名设置流程（创建 DID、添加域名并验证 TXT/NS、创建组织、添加用户、加入组织）。被 `manual/best-practices/set-custom-domain.md`、`manual/larepass/create-org-account.md`、`manual/space/host-domain.md`、`manual/space/manage-domain.md` 引用。
- **openwebui-embeddings.md**：Open WebUI 知识库和网页搜索共用的嵌入配置步骤，使用 `configure-embeddings` 区域。
- **ai-service-connections.md**：仅复用 `router-openai-base-url`（Olares 内应用的 OpenAI 兼容 Base URL）和 `model-context-window`（查看实际上下文大小）。模型选择、模型 ID 和 API 密钥字段写在各篇教程中。Claude Code 使用不带 `/v1` 的地址，外部客户端另按其连接方式配置。URL 片段可紧接有序列表的编号引用，续行使用三个空格缩进。

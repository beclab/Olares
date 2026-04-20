# Reusables（可复用片段）

本目录存放通过 `<!--@include: path/to/reusables/file.md{start,end}-->` 在多个文档中引用的共享内容。

- **local-domain.md**：`.local` 域名说明、URL 格式、HTTP 说明及故障排除（Chrome、Safari）。被 `manual/get-started/local-access.md`、`manual/best-practices/local-access.md` 引用。
- **larepass-vpn.md**：LarePass VPN 步骤（下载、启用、确认连接类型）及常见问题（链接至故障排查文档）。被 `manual/get-started/local-access.md`、`manual/best-practices/local-access.md` 引用。
- **sync-files.md**：同步文件至本地（引言与提示、创建库、开启同步、管理同步）。被 `manual/larepass/manage-files.md`、`manual/olares/files/sync-files.md` 引用。
- **export-system-logs.md**：通过设置 > 高级 > 日志导出系统日志的步骤。被 `manual/olares/settings/developer.md`、`manual/help/request-technical-support.md` 引用。
- **custom-domain.md**：自定义域名设置流程（创建 DID、添加域名并验证 TXT/NS、创建组织、添加用户、加入组织）。被 `manual/best-practices/set-custom-domain.md`、`manual/larepass/create-org-account.md`、`manual/space/host-domain.md`、`manual/space/manage-domain.md` 引用。

在各文件顶部注释中注明可引用的行号范围。

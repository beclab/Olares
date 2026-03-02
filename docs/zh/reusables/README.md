# Reusables（可复用片段）

本目录存放通过 `<!--@include: path/to/reusables/file.md{start,end}-->` 在多个文档中引用的共享内容。

- **local-domain.md**：`.local` 域名说明、URL 格式、HTTP 说明及故障排除（Chrome、Safari）。被 `manual/get-started/local-access.md`、`manual/best-practices/local-access.md` 引用。
- **larepass-vpn.md**：LarePass VPN 步骤（下载、启用、确认连接类型）及常见问题（Mac 扩展重置、Windows 杀毒软件）。被 `manual/get-started/local-access.md`、`manual/best-practices/local-access.md` 引用。

在各文件顶部注释中注明可引用的行号范围。

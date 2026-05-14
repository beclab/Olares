# Reusables

This directory holds shared content included in multiple docs via `<!--@include: path/to/reusables/file.md#region-name-->`.

Add new reusable fragments here and wrap each reusable block in a named VS Code region:

```markdown
<!-- #region region-name -->
Reusable content.
<!-- #endregion region-name -->
```

Use stable, descriptive region names so source line changes do not break references.

- **local-domain.md**: .local domain description, URL format, HTTP note, and troubleshooting (Chrome, Safari). Used by `manual/get-started/local-access.md`, `manual/best-practices/local-access.md`, and `one/access-olares-via-local-domain.md`.
- **larepass-vpn.md**: LarePass VPN procedure (Download, Enable, Verify connection type) and FAQs linking to the troubleshooting doc. Used by `manual/get-started/local-access.md`, `manual/best-practices/local-access.md`, and `one/access-olares-via-vpn.md`.
- **sync-files.md**: Sync files to local (intro, Create a library, Enable synchronization, Manage synchronization). Used by `manual/larepass/manage-files.md` and `manual/olares/files/sync-files.md`.
- **export-system-logs.md**: Steps to export system logs via Settings > Advanced > Logs. Used by `manual/olares/settings/developer.md` and `manual/help/request-technical-support.md`.
- **custom-domain.md**: Custom domain setup procedures (Create DID, Add domain with TXT/NS verification, Create organization, Add user, Join organization). Used by `manual/best-practices/set-custom-domain.md`, `manual/larepass/create-org-account.md`, `manual/space/host-domain.md`, and `manual/space/manage-domain.md`.


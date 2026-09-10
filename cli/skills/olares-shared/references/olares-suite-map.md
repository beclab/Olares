# Suite map — which skill owns a task, and getting the binary

Read this when the task has not been assigned to a skill yet, or when `olares-cli` is not on the machine. Once a domain skill is chosen, its own front door carries everything the work needs.

## Skill suite map

| Skill | Use it for |
|---|---|
| [`olares-shared`](../SKILL.md) | Suite routing, platform entry points, profile/auth decisions |
| [`olares-market`](../../olares-market/SKILL.md) | Install and manage catalog/uploaded apps; lifecycle and chart transfer |
| [`olares-settings`](../../olares-settings/SKILL.md) | Post-install app/system configuration, users, VPN, backup and integrations |
| [`olares-cluster`](../../olares-cluster/SKILL.md) | Runtime objects, logs, jobs, namespaces, nodes and middleware |
| [`olares-dashboard`](../../olares-dashboard/SKILL.md) | CPU, memory, disk, network, pod, GPU and fan metrics |
| [`olares-files`](../../olares-files/SKILL.md) | Browse and modify Drive, Sync, cache, external and cloud files |
| [`olares-knowledge`](../../olares-knowledge/SKILL.md) | URL, yt-dlp, aria2, torrent and Hugging Face download tasks |
| [`olares-search`](../../olares-search/SKILL.md) | Full-content file search and installed-app title search |
| [`olares-router`](../../olares-router/SKILL.md) | Configure, install, call and diagnose models through Router |
| [`olares-chart`](../../olares-chart/SKILL.md) | Author, validate and deploy an app's Olares chart |
| [`olares-publish`](../../olares-publish/SKILL.md) | Prepare and submit a public Olares Market listing |
| [`olares-doctor`](../../olares-doctor/SKILL.md) | Diagnose an app/system runtime failure and route the fix |

Porting and debugging an app commonly combines `chart` (author/fix), `market` (lifecycle), the shared platform model and `doctor` (root-cause diagnosis).

Host installation, node joining, OS upgrades and GPU drivers use the kubeconfig-backed `olares-cli node` / `os` / `gpu` trees, not this profile-backed skill suite.

## If olares-cli is not on PATH

Every command in this suite needs it, and the binary carries this suite, so a machine with the skills but no binary got them from a registry rather than from a release. Install it, then have it write the skills that match itself:

```bash
npm install -g @olares/cli@latest
olares-cli skills install
```

This installs a client, not Olares itself: it operates an existing Olares instance over the network and does not create one. An Olares host already has the binary at `/usr/local/bin/olares-cli` with the host-side `node` / `os` / `gpu` trees this suite does not use.

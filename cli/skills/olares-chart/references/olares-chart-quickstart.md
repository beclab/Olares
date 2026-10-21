# Quickstart: one published image with a web UI

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> Use this path when the upstream already publishes a container image and the app is one web process plus its data directory. A bundled database, a GPU, a headless service or several cooperating images send you back to the full concern router in the parent.

Most of the time on this path goes into reading upstream, not into Olares. Read the image's documentation once, for four facts, before writing anything: the **port** it listens on, the **data directory** it writes, the **environment variables it refuses to start without**, and whether it has its **own login**.

## 1. Image and architecture

```bash
olares-cli cluster node list                  # the target node's arch
docker manifest inspect <image>:<tag>         # must list that arch
```

Pin a release tag, not `latest`. A missing arch is the [image](olares-chart-image.md) path, not this one.

## 2. A compose file to convert

Use upstream's compose file if it has one; otherwise write the minimum:

```yaml
services:
  myapp:
    image: <image>:<tag>
    ports: ["8080:8080"]
    volumes: ["./data:/data"]
    environment:
      LISTEN_ADDR: 0.0.0.0:8080
```

```bash
olares-cli chart from-compose --name myapp -f compose.yml --title "My App"
```

When upstream's file reads secrets as `${VAR}`, add `--no-interpolate` so nothing from your shell is baked in; each `${VAR}` it keeps still needs a real value ([from-compose](olares-chart-from-compose.md)).

## 3. The edits every such app needs

| Edit | Where | Detail |
|---|---|---|
| Data directory → `.Values.userspace.appData` plus `permission.appData: true`; delete the kompose PVC | templates, manifest | [manifest](olares-chart-manifest.md) |
| Process uid, and who owns the data directory | deployment | [run identity](olares-chart-run-as-user.md) |
| Variables the app refuses to start without (secret keys, base URL) | Secret + `env` | [secrets](olares-chart-secrets.md), [environment](olares-chart-env.md) |
| Its own login, when Olares already signs the user in | `env` | [entrances](olares-chart-entrances.md#behind-the-olares-sign-in) |
| How it opens (`openMethod`), title, icon | manifest `entrances[]`, `metadata.icon` | [entrances](olares-chart-entrances.md) |
| `spec.supportArch` with the arch from step 1 | manifest | [image](olares-chart-image.md) |

A required variable that is missing is the most common first crash: the pod restarts while the row still reads `installing`. Set every one of them before the first install.

## 4. Lint, deploy, verify

```bash
olares-cli chart lint ./myapp
olares-cli chart package ./myapp
olares-cli market upload ./myapp-<version>.tgz
olares-cli market install myapp -s upload --version <version> --watch --fail-on-crashloop --watch-timeout 1m -o json
```

From here the loop is [deploy](olares-chart-deploy.md): a crash means read the logs, fix, bump, re-upload and [recover](olares-chart-recover.md). Finish with [verify](olares-chart-verify.md) — the task is done when the app opens from the Desktop, not when the row says `running`.

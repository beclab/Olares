# chart from-compose (scaffold a chart from docker-compose)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli chart from-compose --help`. This file adds what `--help` cannot: prep, the entrance-label trick, and how to read the output before refining.

`from-compose` (alias `init`) runs the same kompose conversion Olares Studio / devbox use, then writes an Olares chart layout. It is **local-only** — no Olares login, no cluster.

```bash
olares-cli chart from-compose --name myapp -f docker-compose.yml
olares-cli chart from-compose --name myapp -f compose.yml -o ./charts/myapp --title "My App"
olares-cli chart from-compose --name myapp -f base.yml -f override.yml      # merged in order
```

## Before you run

- **Every service needs a pullable, target-arch `image:`.** Olares pulls images from a registry and never builds from source, so build-only services (kompose writes them as `image: <service>`, e.g. `image: app` / `image: db`) and wrong-architecture images will fail to deploy. If any service lacks a real, arch-correct image, run the **Image capability** first; if you also lack a usable compose, see the compose-input capability.
- **Pick a valid app name**: `^[a-z][a-z0-9]{0,29}$` (lowercase, starts with a letter, ≤30 chars). It becomes `metadata.name`, `metadata.appid`, the chart name, and the default output dir (`./<name>`).
- **Label the entrance service** in the compose file so the right workload is exposed and renamed to the app name:
  ```yaml
  services:
    web:
      image: ...
      labels:
        olares.service.type: Entrance
      ports: ["8080:80"]
  ```
  Without the label, `from-compose` falls back to the first service that exposes a port, else a `port: 80` placeholder you must fix.

## What each flag controls

| Flag | Effect |
|---|---|
| `-f, --file` (repeatable) | compose file(s); multiple are merged by kompose in order |
| `--name` (required) | app name → `metadata.name`/`appid`, chart name, default output dir |
| `-o, --output` | chart root dir (default `./<name>`) |
| `--title` | human title (default = name) |
| `--type` | `app` (default) / `recommend` / `middleware` |
| `--new-schema` | **deprecated no-op** — the scaffold always emits the canonical `apiVersion: v3` + `olaresManifest.version: 0.12.0` manifest (resources under `spec.accelerator[mode=cpu]`; the flat `spec.requiredCpu/...` envelope is the equivalent no-mode form — see the Accelerator sizing §A.1) |

## Reading the output

The command prints the absolute chart path and a reminder to refine + lint. Then inspect:

- `OlaresManifest.yaml` — the stub you will refine (see the Manifest refinement areas; metadata can stay a stub for local deploy). It already carries the canonical version block (`apiVersion: v3`, `olaresManifest.version: 0.12.0`, and `olares >=1.12.6-0` as a `system` dependency) plus `workloadReplicas` for every rendered Deployment/StatefulSet.
- `templates/deployment-<app>.yaml` — the primary workload (renamed to the app name; required by lint). Its `spec.replicas` is wired to `{{ .Values.workloads.<name>.replicaCount }}` (seeded in `values.yaml`) so app-service can scale it for install / suspend / resume.
- `templates/service-*.yaml` — exposed services; the entrance `host` points at one of these service names.
- `templates/persistentvolumeclaim-*.yaml` — one per compose volume; **these are the storage decisions you must revisit** (most should become userspace volumes; PVCs belonging to a bundled db must be deleted along with that db's workload — see middleware below).

## Conversion limitations to expect

- **`build:`-only services** (no `image:`, or a local-only tag) come out as `image: <service>` — not a pullable reference. These won't deploy; resolve them with the Image capability before scaffolding.
- **`hostPath` / bind mounts** (`./dir:/path`) are dropped by kompose with a warning — the host path won't exist on Olares. Re-model these as userspace volumes.
- **Bundled db/queue services** (`postgres`/`redis`/`mongodb`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats`) come through as plain workloads. **Delete them and wire to system middleware** — do not keep them just because they render (see manifest §3; this is the default, not optional).
- **`depends_on`, healthchecks, restart policies** don't all map 1:1; verify the rendered templates.
- **Workloads you add by hand** (extra Deployments/StatefulSets beyond what kompose rendered) must each be added to `workloadReplicas`, get a `values.yaml` `workloads.<name>.replicaCount`, and wire `spec.replicas: {{ .Values.workloads.<name>.replicaCount }}` — otherwise suspend/resume won't control them (see manifest Workloads & replicas).
- The conversion clears the **local structural `lint`**, but a passing local `lint` is not proof the target Olares accepts it and not proof it is production-ready — confirm `workloadReplicas` and the other required manifest fields yourself (see manifest Workloads & replicas), and the four refinement areas in the parent skill are mandatory before the app will run well. Metadata (§1) can stay a stub for local deploy; functional refine (§2–§4) is always required.
- If a fresh scaffold fails on version fields, do **not** change `OlaresManifest.yaml` to `v1`/`v2`, lower `olaresManifest.version`, or lower the Olares dependency. Check that you are running the current `olares-cli` and current skill. Remember that `Chart.yaml apiVersion: v2` is correct Helm metadata and is independent of `OlaresManifest.yaml apiVersion: v3`.

## Next step

Once refined, validate in a loop:

```bash
olares-cli chart lint ./myapp      # see the Validate-local (lint) step
```

Once `lint` passes, deploy to a real Olares automatically (no extra confirmation needed; proceed unless olares-shared's [auth-readiness gate](../../olares-shared/SKILL.md#auth-readiness-gate) says stop, or a failure is clearly not a chart problem) — the Deploy step. To list it on the public Market afterwards, see [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).

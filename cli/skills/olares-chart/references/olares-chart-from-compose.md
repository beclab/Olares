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
- **Label the entrance service** in the compose file so the right workload is exposed (and, unless another compose service is already named exactly like the app, renamed to the app name):
  ```yaml
  services:
    web:
      image: ...
      labels:
        olares.service.type: Entrance
      ports: ["8080:80"]
  ```
  Without the label, `from-compose` falls back in order: the service fronting a workload already named exactly like the app, else the lowest TCP port among the services that do not look like a datastore (judged by image name or by a well-known port such as 5432 / 3306 / 6379); if every service looks like a datastore it takes the first one exposing a TCP port, and with no candidate at all it writes a `port: 80` placeholder you must fix. The command always prints which service it picked and why — read that line.

## What each flag controls

| Flag | Effect |
|---|---|
| `-f, --file` (repeatable) | compose file(s); multiple are merged by kompose in order |
| `--name` (required) | app name → `metadata.name`/`appid`, chart name, default output dir |
| `-o, --output` | chart root dir (default `./<name>`) |
| `--title` | human title (default = name) |
| `--type` | `app` (default) / `recommend` / `middleware` |
| `--profile` (repeatable) | activate a compose profile |
| `--no-interpolate` | keep `${VAR}` verbatim instead of resolving it from your shell. Without it, a variable that is unset in your environment resolves to an empty string and is baked into the chart |
| `--new-schema` | **deprecated no-op** — the scaffold always emits the canonical `apiVersion: v3` + `olaresManifest.version: 0.12.0` manifest (resources under `spec.accelerator[mode=cpu]`; the flat `spec.requiredCpu/...` envelope is the equivalent no-mode form — see the Accelerator sizing §A.1) |

## Reading the output

The command prints the chart path, the entrance it picked and why, and a `review before deploying:` list of every guess and lossy mapping it made (renamed objects, the workload renamed to the app name, bundled datastores, dropped bind mounts, generated PVCs, services with no pullable image, the restart policies and schedules it had to drop, and any name it could not repair). **That list is the refinement worklist** — work through it rather than rediscovering the same items in the templates. Then inspect:

- `OlaresManifest.yaml` — the stub you will refine (see the Manifest refinement areas; metadata can stay a stub for local deploy). It already carries the canonical version block (`apiVersion: v3`, `olaresManifest.version: 0.12.0`, and `olares >=1.12.6-0` as a `system` dependency) plus `workloadReplicas` for every rendered Deployment/StatefulSet.
- `templates/deployment-<app>.yaml` — the primary workload (the entrance service's workload, renamed to the app name; required by lint). If a compose service is already named exactly like the app, that one keeps the name and nothing is renamed, so the entrance may front a different workload — the notice list says so when it happens. Its `spec.replicas` is wired to `{{ .Values.workloads.<name>.replicaCount }}` (seeded in `values.yaml`) so app-service can scale it for install / suspend / resume. A workload whose name contains a dash uses the equivalent `{{ (index .Values.workloads "<name>").replicaCount }}`, because Helm cannot reach a dashed key with dotted syntax.
- `templates/service-*.yaml` — exposed services; the entrance `host` points at one of these service names.
- `templates/persistentvolumeclaim-*.yaml` — one per compose volume; **these are the storage decisions you must revisit** (most should become userspace volumes; PVCs belonging to a bundled db must be deleted along with that db's workload — see middleware below).

## Conversion limitations to expect

- **`build:`-only services** (no `image:`, or a local-only tag) come out as `image: <service>` — not a pullable reference. These won't deploy; resolve them with the Image capability before scaffolding.
- **`hostPath` / bind mounts** (`./dir:/path`) never survive as mounts: kompose drops a missing or empty host path and gives you an empty PVC instead, **copies** an existing file or non-empty directory into a ConfigMap inside the chart, and skips socket paths (`docker.sock`) entirely. The notice list says which of the three happened per mount. Re-model them as userspace volumes; a host socket has no equivalent at all.
- **A compose *named* volume loses whatever the image baked at that path.** Docker seeds a named volume from the image the first time it is used; a `hostPath` or PVC does not — it shadows the image's directory with an empty one. An image that ships its web root, plugin set or default config *inside* the mounted path therefore comes up empty: the app runs and looks healthy, but answers 403/404, shows no plugins, or walks through first-run setup again, with nothing in the logs saying why. Seed it from an initContainer running the **app's own image** with the volume mounted somewhere else, so the image's copy is still visible to copy from — `sh -c '[ -z "$(ls -A /seed)" ] && cp -a /var/www/html/. /seed/'`. That is also where the uid-1000 ownership prep belongs — non-recursively (see the run identity guidance).
- **Bundled db/queue services** (`postgres`/`redis`/`mongodb`/`mysql`/`mariadb`/`minio`/`rabbitmq`/`nats`) come through as plain workloads. **Delete them and wire to system middleware** — do not keep them just because they render (see manifest §3; this is the default, not optional).
- **`depends_on` and healthchecks** don't all map 1:1; verify the rendered templates.
- **`restart: no` / `on-failure`, `deploy.mode: global` and `kompose.cronjob.schedule` are dropped by default** — kompose would emit a bare Pod, a DaemonSet and a CronJob, none of which Olares can scale for install / suspend / resume, so the service comes out as a replica-controlled workload with `restartPolicy: Always`. A `kompose.controller.type` label changes which kind that is without bringing the restart or schedule semantics back (the notice list names the kind each service actually became); the one exception is `deploy.mode: global` with `kompose.controller.type: daemonset`, which really does render a DaemonSet and is not reported. A one-shot or scheduled job therefore becomes a long-running workload — rethink it rather than shipping it. If the app genuinely needs periodic work, either schedule it inside the container or add a CronJob template by hand and keep it out of `workloadReplicas`.
- **Some compose service names cannot be repaired at all** and are reported instead: a name starting with a digit (`1web`) is not a valid Service name, and a name starting or ending with `.` / `_` / `-` (`_web`, which kompose turns into the selector value `-web`) or longer than 63 characters is not a valid label value. Both end up in the pod labels kompose selects on, which the conversion cannot rewrite without breaking the selector, so **local lint still passes and the install fails** — rename the compose service.
- **Compose names are normalized** to valid Kubernetes names (`web_app` → `web-app`, `web.ui` → `web-ui`, volume `PGData` → `pgdata`), and the references that point at them (`claimName`, `configMapRef`, pod volume and container names, a StatefulSet's governing `serviceName`, and any Ingress backend) are normalized with them. A hostname hard-coded in another service's `environment` or `command` is **not** rewritten and still points at the old name; fix those by hand. If two of the objects kompose emits normalize onto one name (two volumes of one service, say), the conversion errors out instead of letting one template overwrite the other. Two compose **service** names that collapse onto one (`web_app` and `web-app`) are out of its reach: kompose merges them while loading the compose file, so one service loses its ports and its workload with nothing reported, and which one survives changes between runs — never give two services names that differ only in `.` / `_` / `-`.
- **The workload kind is pinned to Deployment**; a service-level `kompose.controller.type` label is the only way to change it, and only the exact lowercase values `deployment` / `statefulset` / `daemonset` are understood. `statefulset` still lands in `workloadReplicas`, so suspend/resume keeps controlling it; `daemonset` does not (DaemonSets are one-per-node — see manifest Workloads & replicas). A `statefulset` on a service with no `ports` is reported: kompose points the StatefulSet at a governing Service it only creates for a service that has ports, so its pods get no resolvable names. Any other value (`StatefulSet`, `replicaset`, an empty value) leaves that service with **no workload at all** — its Service, PVCs, `env_file` ConfigMaps and any Ingress are still emitted, so the chart renders with a Service selecting nothing.
- **The HorizontalPodAutoscaler kompose renders from `kompose.hpa.*` labels is dropped** and reported: it would drive the same `spec.replicas` as `workloadReplicas`, which is how Olares installs, suspends and resumes an app, and lint requires `workloadReplicas` to match the rendered workloads exactly, so there is no placement that works. The service's workload still comes out replica-controlled through `workloadReplicas`, with an initial replica count of `1`. An app that genuinely needs autoscaling needs a mechanism outside Olares' replica control.
- **A compose file that renders no Deployment/StatefulSet** (e.g. every service carries `kompose.controller.type` pointing elsewhere) is rejected with an error instead of producing an unusable chart: Olares has nothing to scale.
- **Workloads you add by hand** (extra Deployments/StatefulSets beyond what kompose rendered) must each be added to `workloadReplicas`, get a `values.yaml` `workloads.<name>.replicaCount`, and wire `spec.replicas: {{ .Values.workloads.<name>.replicaCount }}` — otherwise suspend/resume won't control them (see manifest Workloads & replicas).
- The conversion clears the **local structural `lint`**, but a passing local `lint` is not proof the target Olares accepts it and not proof it is production-ready — confirm `workloadReplicas` and the other required manifest fields yourself (see manifest Workloads & replicas), and the four refinement areas in the parent skill are mandatory before the app will run well. Metadata (§1) can stay a stub for local deploy; functional refine (§2–§4) is always required.
- If a fresh scaffold fails on version fields, do **not** change `OlaresManifest.yaml` to `v1`/`v2`, lower `olaresManifest.version`, or lower the Olares dependency. Check that you are running the current `olares-cli` and current skill. Remember that `Chart.yaml apiVersion: v2` is correct Helm metadata and is independent of `OlaresManifest.yaml apiVersion: v3`.

## Next step

Once refined, validate in a loop:

```bash
olares-cli chart lint ./myapp      # see the Validate-local (lint) step
```

Once `lint` passes, deploy to a real Olares automatically (no extra confirmation needed; proceed unless olares-shared's [auth-readiness gate](../../olares-shared/SKILL.md#auth-readiness-gate) says stop, or a failure is clearly not a chart problem) — the Deploy step. To list it on the public Market afterwards, see [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).

# Typical assembly & conversion output

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This is the end-to-end overview the SKILL routes to: one way to compose the capabilities, and what `from-compose` actually emits. Not a fixed pipeline — start wherever your state tables put you and loop across the coupling edges as failures surface.

## A typical assembly (compose with a build-only image)

```
 Packaging (agent-driven; only if an image is missing / wrong-arch):
 P1. docker?    docker version && docker buildx version   # else guide install
 P2. registry   ask which registry + <user>/<repo>; check login — if not authed, the developer runs `docker login` (agent can't type their token)
 P3. build+push docker buildx build --platform linux/<node-arch> -t <ref>:<tag> --push <ctx>
                -> you run build+push, then wire <ref>:<tag> into every build-only `image:` in the compose

 Deployment authoring (no login):
 D1. scaffold   olares-cli chart from-compose --name <app> -f docker-compose.yml
 D2. refine     edit OlaresManifest.yaml + templates/ for the 4 refinement areas
                (metadata stub OK for local deploy — see the Manifest refinement areas)
 D3. lint       olares-cli chart lint ./<app>        # loop D2<->D3 until OK
 D4. package    bump metadata.version (= Chart.yaml version), then olares-cli chart package ./<app>

 Deploy to your Olares (requires login; automatic once lint passes):
 V1. usable?    olares-cli profile list              # apply olares-shared auth-readiness gate: go (logged-in/expired) or stop (invalidated/never -> tell developer)
 V2. auto       after lint passes + gate=go, run V3-V6 WITHOUT asking (stop only if the gate says stop, or a failure is clearly not a chart problem)
 V3. upload     olares-cli market upload ./<app>-<ver>.tgz
 V4. run        olares-cli market install <app> -s upload --version <ver> --watch -o json
 V5. on failure diagnose via olares-doctor (symptom -> root cause), then loop back
 V6. decide     loop back: chart problem -> D2 ; image problem -> P3 ; uid/EACCES -> run identity (uid 1000) guidance ; not a chart problem -> break out, report & ask
 V7. cleanup    olares-cli market uninstall <app> --watch ; olares-cli market delete <app>
```

Step D1 produces a chart that **already passes `lint`** but is NOT yet a good app: kompose translates containers literally and cannot make product decisions. The value you add is D2. Treat the generated `OlaresManifest.yaml` as a stub — for deploying to your own Olares, §1 Metadata can stay a stub. The V steps cross into sibling skills — full procedure in the Deploy step; runtime diagnosis is [`../../olares-doctor/SKILL.md`](../../olares-doctor/SKILL.md). Drive the loop automatically once `lint` passes (proceed unless olares-shared's [auth-readiness gate](../../olares-shared/SKILL.md#auth-readiness-gate) says stop, or the failure is clearly not a chart problem); never log in on the developer's behalf without asking.

> **Publishing to the public Market** (multi-arch build, full market-ready metadata, the `beclab/apps` PR, paid apps) is the [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md) skill — start there once the app runs on your Olares.

## What the conversion produces

```
<output>/
├── Chart.yaml              # helm chart metadata (name/version pinned to 0.0.1)
├── OlaresManifest.yaml     # Olares app manifest — the file you refine
├── values.yaml             # seeded with workloads.<name>.replicaCount; add more as you template values
└── templates/
    ├── deployment-<app>.yaml          # the primary workload, renamed to <app>
    ├── deployment-<svc>.yaml          # one per extra compose service
    ├── service-<svc>.yaml             # one per exposed compose service
    └── persistentvolumeclaim-*.yaml   # one per named/anonymous compose volume
```

- Every resource is namespaced with `namespace: '{{ .Release.Namespace }}'`.
- Default CPU/memory requests+limits are stamped onto every container.
- One **entrance** is auto-detected (the `olares.service.type: Entrance`-labeled service, else the first service with a port, else a `port: 80` placeholder).
- The scaffold emits the complete canonical version combination: `OlaresManifest.yaml apiVersion: v3`, `olaresManifest.version: 0.12.0`, and `options.dependencies` `olares >=1.12.6-0` (`type: system`). `Chart.yaml apiVersion: v2` remains Helm metadata and is intentionally different.
- The scaffold also emits `workloadReplicas.<workload>: 1` (with the matching `values.yaml` `workloads.<name>.replicaCount`, and each workload's `spec.replicas` wired to `{{ .Values.workloads.<name>.replicaCount }}` so suspend/resume work), so a fresh scaffold passes `lint`.
- Never respond to a version-related lint failure by downgrading the canonical combination to manifest v1/v2 or Olares `<1.12.6`; check for an old CLI or old skill instead.

> **Tip:** label the service you want exposed in the compose file with `labels: { olares.service.type: Entrance }` so the right workload becomes the entrance and gets renamed to the app name.

# Typical assembly & conversion output

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This is the end-to-end overview the SKILL routes to: one way to compose the capabilities, and what `from-compose` actually emits. Not a fixed pipeline — start wherever your state tables put you and loop across the coupling edges as failures surface.

## A typical assembly (compose with a build-only image)

```
 0. target       local-run | market-distribute   # gates P3 arch flags + D2 metadata depth

 Packaging (agent-driven; only if an image is missing / wrong-arch):
 P1. docker?    docker version && docker buildx version   # else guide install
 P2. registry   ask which registry + <user>/<repo>; check login — if not authed, the developer runs `docker login` (agent can't type their token)
 P3. build+push local-run:     docker buildx build --platform linux/<node-arch> -t <ref>:<tag> --push <ctx>
                market:        docker buildx build --platform linux/amd64,linux/arm64 -t <ref>:<tag> --push <ctx>
                -> you run build+push, then wire <ref>:<tag> into every build-only `image:` in the compose

 Deployment authoring (no login):
 D1. scaffold   olares-cli chart from-compose --name <app> -f docker-compose.yml
 D2. refine     edit OlaresManifest.yaml + templates/ for the 4 refinement areas
                (metadata depth per release target — see manifest.md)
 D3. lint       olares-cli chart lint ./<app>        # loop D2<->D3 until OK
 D4. package    olares-cli chart package ./<app>

 Publish-local (requires login + developer consent; local-run done here):
 V1. logged-in? olares-cli profile list              # if not: tell developer, stop
 V2. ask        confirm with the developer before uploading to a real Olares
 V3. upload     olares-cli market upload ./<app>-<ver>.tgz
 V4. run        olares-cli market install <app> -s upload --version <ver> --watch -o json
 V5. on failure fetch market / chartrepo / app-service / app-pod logs and diagnose
 V6. decide     loop back: chart problem -> D2 ; image problem -> P3 ; uid/EACCES -> run-as-user.md ; else report & ask
 V7. cleanup    olares-cli market uninstall <app> --watch ; olares-cli market delete <app>

 Publish-market (market-distribute only; requires V pass first):
 M1. polish     full metadata, categories, listing images, spec.supportArch — publish-targets checklist
 M2. lint       olares-cli chart lint ./<app>        # re-check after polish
 M3. package    olares-cli chart package ./<app>
 M4. fork       developer forks beclab/apps, adds OAC + owners file
 M5. PR         [NEW][<app>][<ver>] Summary — see market-submit.md
 M6. wait        GitBot validates → auto-merge → app appears in Market
```

Step D1 produces a chart that **already passes `lint`** but is NOT yet a good app: kompose translates containers literally and cannot make product decisions. The value you add is D2. Treat the generated `OlaresManifest.yaml` as a stub — how much you polish §1 Metadata depends on the release target. The V steps cross into sibling skills — full procedure in [olares-chart-publish-verify.md](olares-chart-publish-verify.md). M steps are for market-distribute only — [olares-chart-market-submit.md](olares-chart-market-submit.md). Only proceed past D3 upload with the developer's consent.

## What the conversion produces

```
<output>/
├── Chart.yaml              # helm chart metadata (name/version pinned to 0.0.1)
├── OlaresManifest.yaml     # Olares app manifest — the file you refine
├── values.yaml             # empty; fill if you template values
└── templates/
    ├── deployment-<app>.yaml          # the primary workload, renamed to <app>
    ├── deployment-<svc>.yaml          # one per extra compose service
    ├── service-<svc>.yaml             # one per exposed compose service
    └── persistentvolumeclaim-*.yaml   # one per named/anonymous compose volume
```

- Every resource is namespaced with `namespace: '{{ .Release.Namespace }}'`.
- Default CPU/memory requests+limits are stamped onto every container.
- One **entrance** is auto-detected (the `olares.service.type: Entrance`-labeled service, else the first service with a port, else a `port: 80` placeholder).
- `olaresManifest.version` is `0.8.0` (legacy) unless you pass `--new-schema` (`0.12.0`, resources under `spec.accelerator`) — see [olares-chart-manifest.md](olares-chart-manifest.md).

> **Tip:** label the service you want exposed in the compose file with `labels: { olares.service.type: Entrance }` so the right workload becomes the entrance and gets renamed to the app name.

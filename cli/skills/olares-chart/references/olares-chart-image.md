# Image: a pullable, arch-correct image (packaging axis)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This is the **packaging** capability. It is orthogonal to having a compose — a compose says nothing about whether an image needs building — and can be entered up front or looped back to later (an install that hits `ImagePullBackOff` sends you here).

## Arch strategy

Deploying to your Olares only needs the **target Olares node's arch** (single-arch) — query it with `olares-cli cluster node list` (needs login). The development host may have a different architecture, so never derive the image platform from `uname`, `runtime.GOARCH`, or Docker's default platform. Multi-arch (`linux/amd64,linux/arm64` + a matching `spec.supportArch: [amd64, arm64]`) is only required when **publishing to the public Market** — see the [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md) skill.

Declare that same architecture in `spec.supportArch`: it is **required and must be non-empty** (`lint` rejects a missing or empty list), and a single entry also makes app-service pin the pods to matching nodes via a `kubernetes.io/arch` nodeSelector. It remains a **declaration about the chart** — nothing opens the image to check — so a chart declaring `amd64` around an arm64 image passes every check and still crashes.

## When you need it

Olares installs apps by **pulling images from a registry; it never builds from source.** So every workload must reference an image that is publicly pullable **for the target node's architecture**. Skip this capability only when every service already does.

- **Already pullable + arch-correct:** nothing to do for that service.
- **Repo has no Dockerfile:** read the code to infer the runtime (language, start command, listening port, required env, data directories), **author a Dockerfile**, then build+push.
- **Repo has a Dockerfile but no official (or no target-arch) image:** build+push from the Dockerfile.

## The image-readiness gate

For an image **you** build, three properties stay invisible until the cluster rejects it: the architecture actually built, whether it can be pulled **without credentials**, and whether the process starts at all. Run each step and read its output — that is the difference between a defect found here in seconds and one found a deploy cycle later.

| # | Step | The assertion |
|---|---|---|
| 1 | Resolve the target arch (`olares-cli cluster node list`) | a value exists, and it came from the target rather than from `uname` |
| 2 | `docker buildx build --platform linux/<target-arch> --load -t <ref>:<tag> <ctx>` | the build carries an explicit platform |
| 3 | `docker image inspect <ref>:<tag> --format '{{.Architecture}}'` | the output **equals** the target arch |
| 4 | Start the container (when that proves something — see below) | the process does not exit immediately; an HTTP service answers on its declared port |
| 5 | `docker buildx build ... --push` (or `docker push <ref>:<tag>`) | the push reports success for that exact tag |
| 6 | Inspect anonymously, with an empty `DOCKER_CONFIG` | the registry serves that tag to a caller holding no credentials |

Step 6 exists because Olares nodes pull anonymously and a logged-in shell cannot tell a public repository from a private one. It proves access, not architecture: a single-platform manifest does not always carry a `platform` field, which is why step 3 asserts the arch locally.

A reused tag is the fourth failure mode and the one no check here can see — a node keeps serving the layers it already cached. Hence the rule under Hard rules: every rebuild gets a new tag. Do not wire an image into the chart until step 6 passes.

### Steps 2-3: build locally first, then assert the architecture

`--push` on its own leaves nothing behind to inspect, so load the image and assert before publishing it:

```bash
docker buildx build --platform linux/<target-arch> --load -t <ref>:<tag> <build-context>
docker image inspect <ref>:<tag> --format '{{.Architecture}}'   # must print <target-arch>
```

A mismatch means `--platform` was wrong or missing: rebuild, and do not push. (Publishing to the public Market builds multi-arch and cannot use `--load` — that path is [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).)

### Step 4: start it, when starting it proves something

For an ordinary long-running service, a container that exits at once is a defect visible in seconds — a missing entrypoint dependency, an unwritable runtime path ([run-as-user.md](olares-chart-run-as-user.md)), a wrong `CMD`:

```bash
name=<app>-smoke-$$
if docker run -d --name "$name" -p <host>:<container> <ref>:<tag>; then
  sleep 5
  running=$(docker inspect --format '{{.State.Running}}' "$name")
  docker logs "$name"                              # then curl the declared port for an HTTP service
  docker rm -f "$name"
  test "$running" = true                           # assert only after preserving logs
fi
```

No `--rm`: automatic removal deletes the container before `docker logs` can explain a failed start. Unique name behind an `if`: a fixed name that already exists fails the `run` and would send `docker rm -f` at somebody else's container. Emulating a foreign architecture is slow and sometimes impossible — if it cannot run here, say so rather than skipping in silence.

**Skip the port probe — and say why — when the workload cannot answer one here:** GPU / accelerator images, images bound to host devices, anything needing cluster middleware (Postgres, Redis, an `.Values.olaresEnv` value) to finish booting, and jobs that are *supposed* to exit. The deploy loop is their real test.

### Step 6: verify the pull the node will actually make

```bash
(
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  DOCKER_CONFIG="$tmp" docker manifest inspect <ref>:<tag>
)
```

**Never `docker logout`** to simulate an anonymous client — that destroys credentials the developer has to retype. `denied` / `unauthorized` means the repository is private (ghcr defaults to private on first push: set the package visibility to public) or the tag was never pushed. Absence of `platform.architecture` is not a failure: a single-image manifest need not carry it.

## Resolving the target architecture

A wrong-architecture image installs but never runs (`ImagePullBackOff` with `no match for platform`, or the container `exec format error`-crashes). Gate step 1 is where that value comes from:

```bash
olares-cli cluster node list          # ARCHITECTURE shows amd64 / arm64 (needs login)
```

If more than one architecture is listed, identify the node that will run the workload and confirm it with `olares-cli cluster node get <name>` — a single-arch app is pinned to matching nodes, so on a mixed cluster that choice also decides where it can run. Never fall back to the development host's architecture; if the target cannot be identified, stop and ask.

**For an image you did not build**, read its platforms before trusting it — `docker manifest inspect <image-ref>`, or the registry manifest list over HTTP with no docker daemon. One that already covers the target arch needs no gate run; one that doesn't sends you back to building your own.

## GPU / CUDA images

Building a CUDA image (no GPU needed on the build box, custom-kernel arch flags, the amd64 / `nvidia`-mode constraint) and provisioning model weights (initContainer + shared Hugging Face cache) are covered in the GPU / models capability.

## Registry + build/push (agent-driven)

You drive this end to end. The **only** manual step is the developer typing a registry token into `docker login`, and only when they are not already authenticated. **Never invent/hardcode tokens or push under an account the developer didn't choose.**

1. **Resolve `<target-arch>` first** (gate step 1, plus the multi-node rule above).

2. **Ask which registry the developer uses + the target `<user>/<repo>`** (don't assume one):
   - **Docker Hub** — image ref `<dockerhub-user>/<repo>`
   - **GitHub Container Registry (ghcr)** — image ref `ghcr.io/<owner>/<repo>`
   > An Olares-local private registry is not supported here — the image must live on a registry the Olares node can pull from publicly.

3. **Check docker is usable:**
   ```bash
   docker version          # must show a Server section; if it errors, the daemon isn't running
   docker buildx version   # buildx is needed for the explicit --platform build
   ```
   If docker is missing or the daemon is down, point the developer to install / start it: Docker Desktop on macOS/Windows, or the engine on Linux — https://docs.docker.com/get-docker/ . Stop and wait until `docker version` shows a Server.

4. **Check whether they're already logged in to that registry** — don't ask for a login they already have:
   ```bash
   docker login <registry>   # already authed? prints "Authenticating with existing credentials" / "Login Succeeded"
   ```
   Or read `~/.docker/config.json` `auths` for the registry key (Docker Hub → `https://index.docker.io/v1/`, ghcr → `ghcr.io`; a `credsStore`/`credHelpers` entry can be empty but present). A push failing with `unauthorized` / `denied` is the authoritative "not logged in / wrong account" signal. Already logged in → go to step 5. Otherwise ask them to run it — it needs their secret token:
     - Docker Hub: `docker login` with an **access token** (Account Settings → Security → New Access Token).
     - ghcr: `docker login ghcr.io -u <github-user>` with a **GitHub PAT** holding `write:packages`. After the first push, set the package **visibility to public** so Olares can pull it without auth.

5. **Run gate steps 2-6** against `<registry-ref>:<tag>`, once the developer has confirmed that ref. `<build-context>` can be a local path (`.`) or a git URL (e.g. `https://github.com/org/repo.git#main`). Publishing to the public Market instead? Build multi-arch — `--platform linux/amd64,linux/arm64` — per [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md).

## Handoff: wire the image into the compose

Once the gate has passed for every service, replace each `build:` block in the compose (and any local-only `image:` tag like `image: app`) with the pushed `<registry-ref>:<tag>`. Every service is now proven pullable and arch-correct, so proceed to scaffold:

```bash
olares-cli chart from-compose --name <app> -f docker-compose.yml
```

Then continue with the four refinement areas (the Manifest refinement areas) and `chart lint`.

## Run identity (UID/GID 1000)

Olares userspace volumes expect the app process as **uid/gid 1000**, which is a
choice you make while picking or building the image, not only afterwards. Probe
a candidate third-party image with `docker inspect <ref> --format '{{.Config.User}}'`
before wiring it in: an image that stays root needs chart-side work, and one
that cannot be made to run as 1000 needs replacing. What that work is — in the
Dockerfile, the manifest, or an initContainer — is the run identity (uid 1000)
guidance, which owns the decision tree.

## Hard rules

- **Every service must reference a publicly pullable image** for the node arch — no `build:`, no local-only tags, no private registry (until Olares-local registry support lands).
- **Deploy to your Olares:** the image's built architecture must equal the node's. Assert it (gate step 3) rather than inferring it from the chart's `spec.supportArch`, which is a declaration and not a fact about the image. (Multi-arch is only for publishing — [`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md); `spec.supportArch` itself is required either way.)
- **Never bake registry credentials into the chart** (no `imagePullSecrets` with inline tokens, no secrets in `values.yaml`). Public images only.
- **Pin every image to a specific version tag** — **never `:latest`** or an untagged image (implicit `latest`), which drifts and makes installs non-reproducible. `lint` does not enforce this.
- **Bump the tag on every rebuild.** A node that already pulled `<ref>:<tag>` keeps serving the cached layers, so new bytes under a seen tag change nothing there. A fix that "did not take effect" is usually this, not the fix.

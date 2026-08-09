---
outline: [2, 3]
description: Build a component from the Olares repo locally, side-load it onto a running Olares, point the workload at it, and revert when you are done.
head:
  - - meta
    - name: keywords
      content: Olares, development, dev deploy, side-load image, containerd, set-image, contributing
---

# Test a locally built component on a running Olares

When you change a component in this repo — `app-service`, `bfl`, `l4-bfl-proxy`, anything under [`framework/`](https://github.com/beclab/Olares/tree/main/framework) or [`platform/`](https://github.com/beclab/Olares/tree/main/platform) — the released chart still pins a published image tag such as `beclab/app-service:0.6.22`. `olares-cli release` packages charts and manifests but never builds images, so there is no path from "I changed some Go" to "it is running on my Olares" without help.

`olares-cli dev` is that path:

```bash
make dev-deploy C=app-service
```

That builds the component's image, side-loads it into your node's containerd, repoints every workload that references the released image, and waits for the rollout. When you are done:

```bash
make dev-revert C=app-service
```

## Before you begin

- **Go 1.25+** and **docker** (or podman) on the machine you build from.
- A running Olares you are logged into: `olares-cli profile current` should show your Olares ID.
- A way to reach the node's image store — either you are working *on* the node, or you can SSH to it. See [Transports](#transports).

::: warning Not for production instances
`dev deploy` repoints real workloads on a real instance. Side-loaded images live only in the node's image store, and the pull policy is set to `IfNotPresent` so the kubelet does not go looking for them in a registry. That combination is exactly what you want while iterating and exactly what you do not want on an instance you depend on.
:::

## The loop

### One command

```bash
make dev-list                  # what can be dev-built
make dev-deploy C=app-service  # build + push + repoint + wait
make dev-status                # what is currently overridden
make dev-revert C=app-service  # put it back
```

`make dev-deploy` is the three CLI verbs below in sequence. Reach for them directly when you want to build once and deploy several times, or when the component is not in the map.

### The verbs underneath

```bash
docker build -t beclab/app-service:dev -f framework/app-service/Dockerfile framework/app-service
olares-cli dev push beclab/app-service:dev
olares-cli dev deploy beclab/app-service:dev --replaces beclab/app-service --watch
```

`push` moves bytes; `deploy` repoints workloads. They are separate because their failure modes are unrelated — a failed push is a transport or permission problem on the node, a failed deploy is an API or RBAC problem — and because one push often feeds several deploys.

`--replaces` takes the repository the chart references, without a tag. The lookup normalizes tags and digests, so it finds whichever released tag is actually deployed. It is a full-cluster scan by design: a paged subset could miss a reference and leave you with a half-repointed component.

Every target is listed and confirmed once before anything is patched:

```
About to repoint 1 container(s) at beclab/app-service:dev:

NAMESPACE     KIND         NAME         CONTAINER    CURRENT
os-framework  StatefulSet  app-service  app-service  beclab/app-service:0.6.22

Repoint these 1 container(s)? Pods will be recreated [y/N]:
```

Pass `--yes` to skip the prompt in a script. Non-interactive callers must pass it — the verb refuses to fan out unconfirmed when stdin is not a terminal.

### Watching it come up

```bash
olares-cli cluster workload rollout-status os-framework/app-service --kind sts --watch
olares-cli cluster pod logs -f os-framework/<pod>
```

## Transports

`dev push` has to get the image into the containerd namespace the kubelet reads (`k8s.io`). How depends on where you are:

| `--transport` | What it does | When `auto` picks it |
|---------------|--------------|----------------------|
| `local` | `docker save` piped into `ctr images import` on this machine | The CLI is running on the Olares node |
| `ssh` | The same pipeline streamed to the node over SSH, no temp file on either side | A node is configured for the active profile |
| `registry` | Not implemented — push to a registry the node can pull from and use `olares-cli cluster workload set-image` | — |
| `api` | Not implemented — needs a backend that can import an uploaded image | — |

`auto` (the default) tries `local`, then `ssh`, and explains what each would need if neither works.

::: tip A local containerd is not enough
`local` requires evidence of a kubelet on the machine, not just a containerd socket. A developer laptop running Docker has a containerd at the same path, and importing into it would appear to succeed while putting the image somewhere no kubelet will ever look.
:::

### Configuring SSH

```bash
olares-cli dev node set --address 192.168.1.42 --user ronny
olares-cli dev node show
```

The settings are stored per profile, so switching profiles switches nodes. There is no password option: use a key (`--private-key`) or an SSH agent. The account needs to reach containerd, which means root or a sudoer.

## Reverting

```bash
olares-cli dev status
olares-cli dev revert --image beclab/app-service:dev --watch
olares-cli dev revert os-framework/app-service --kind sts   # or name one workload
```

`deploy` records the image *and* the pull policy it replaced in a `dev.olares.io/previous-images` annotation, and `revert` restores both. Restoring the policy matters: `deploy` forces `IfNotPresent` so a side-loaded build is usable at all, and leaving that behind would pin the workload to whatever happens to be cached on the node and quietly defeat the next chart upgrade.

Overriding the same container twice keeps the *first* recorded original, so a chain of dev builds still reverts to the released image rather than to an earlier dev build.

## When something goes wrong

**The pod will not start and nothing looks wrong in the chart.** Run `olares-cli dev status`. A side-loaded image exists only in the node's image store; an image prune deletes it, and with `IfNotPresent` there is no registry to recover it from. Re-run `dev push`.

**`exec format error` in the container log.** The image was built for the wrong architecture. `dev push` compares the image against the cluster's nodes and refuses before transferring, but the check is skipped when it cannot reach the cluster or read the local image. Rebuild with `docker build --platform linux/<arch>`, or through the Makefile:

```bash
make dev-deploy C=app-service DOCKER_BUILD_FLAGS="--platform linux/arm64"
```

**`ctr import` fails with a permission error.** containerd's socket is root-owned. Run as root on the node, or configure passwordless sudo for the SSH account.

**`no such image`.** The build did not produce the tag `push` was asked for. `make dev-deploy` keeps the two in sync; when driving the verbs by hand, check that `docker build -t` and `dev push` name the same tag.

## The component map

[`build/dev-components.yaml`](https://github.com/beclab/Olares/blob/main/build/dev-components.yaml) maps a component name to its build context, Dockerfile, and released image repository. It mirrors the `.github/workflows/module_*_publish_*.yaml` files, which remain the source of truth for what actually gets published.

```bash
olares-cli dev components list
olares-cli dev components get app-service
olares-cli dev components validate    # or: make dev-validate
```

`validate` checks that every context and Dockerfile still exists, and runs as part of the test suite. The publish workflows only run on release, so without it a Dockerfile that moves during a refactor could leave the map pointing at nothing for months.

If your component is not in the map, add it — or drive the verbs directly with your own `docker build`.

## Repointing without the dev loop

`dev deploy` is a wrapper around a verb you can use on its own, including for images that have nothing to do with this repo:

```bash
olares-cli cluster workload set-image os-framework/app-service \
    --kind sts --image ghcr.io/me/app-service:pr-42 --watch
```

It patches one container by name (strategic merge, so sibling containers survive), records the original for `dev revert`, and works the same whether the CLI runs on the node or on your laptop.

## Learn more

- [Olares CLI overview](../cli-overview.md): the three modes and how each authenticates.
- [Olares repository structure](./olares.md): what lives where, and how to build a release from source.

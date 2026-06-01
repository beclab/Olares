# Compose: obtain or author a docker-compose (deployment-input axis)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> `chart from-compose` needs a docker-compose as input. This is the **deployment-input** capability: get a compose, then hand it to [olares-chart-from-compose.md](olares-chart-from-compose.md). It is orthogonal to packaging — see [olares-chart-image.md](olares-chart-image.md) for making each service's image pullable.

You either already have a compose, are handed several, or have only source:

- **Repo ships compose(s):** pick the most Olares-compatible variant (below).
- **No compose (source only):** author one from the code (below).

## Choosing among multiple upstream composes

Projects often ship several composes (dev, full, minimal, ha, ...). Prefer the one that is:

- **single-host** (no Swarm `deploy.replicas` clusters, no multi-node assumptions);
- built on **published images**, not `build:` (or minimally so — you only have to build the few that are build-only, via [olares-chart-image.md](olares-chart-image.md));
- **without** `privileged`, `network_mode: host`, `cap_add`, or host-device mounts (Olares won't allow these);
- explicit about **ports, volumes, env** so the four judgment calls are tractable.

A "self-host" / "production" compose is usually closer to Olares than a "dev" compose (dev composes tend to be all-`build:` from source — exactly what you can't deploy as-is).

## Author a compose from the code

When there is no compose, read the code and write a minimal one. For each service capture:

- **image** — the published, arch-correct image (build+push it first if missing — [olares-chart-image.md](olares-chart-image.md));
- **ports** — the port(s) the process listens on (becomes the entrance / service);
- **environment** — required config/secrets the app reads at startup;
- **volumes** — directories the app writes to and must persist;
- **depends_on** — backing services (db/cache/queue) it needs — usually replaced later by Olares system middleware.

```yaml
services:
  app:
    image: <user>/<repo>:<tag>     # pullable + arch-correct (see olares-chart-image.md)
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://user:pass@db:5432/app
    volumes:
      - app-data:/var/lib/app
    depends_on:
      - db
  db:
    image: postgres:16             # likely dropped for Olares system middleware later
    volumes:
      - db-data:/var/lib/postgresql/data
volumes:
  app-data:
  db-data:
```

Keep it faithful to how the app actually runs; the refinement (userspace volumes, system middleware, entrances) happens after conversion, not here.

## Next step

```bash
olares-cli chart from-compose --name <app> -f docker-compose.yml   # olares-chart-from-compose.md
```

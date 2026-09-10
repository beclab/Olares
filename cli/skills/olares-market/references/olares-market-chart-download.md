# market download (pulling a stored chart back out)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli market download --help`. Pushing one in is [market upload](olares-market-chart-publish.md).

```bash
olares-cli market download mychart                     # ./mychart-1.0.0.tgz
olares-cli market download mychart ./charts/           # into a directory, server-chosen name
olares-cli market download mychart ./mychart.tgz        # exact local filename
olares-cli market download mychart --version 1.0.0     # a specific stored version
olares-cli market download mychart -s market.olares    # read another source
olares-cli market download mychart -o json             # {app, source, version, path, bytes}
```

- **The read side of `upload`.** It serves the exact `.tgz` the market hands the installer, so a chart whose only remaining copy is on the Olares — local working copy deleted, or the upload happened from another machine — can be recovered, edited, re-packaged and re-uploaded.
- **It reads, so it takes a source.** Unlike `upload` and `delete`, which pin the bucket to `upload`, pointing `download` at another source cannot desynchronize anything. It defaults to `upload` and accepts `-s <id>`.
- **`--version` omitted → whatever version the market currently holds.** Resolution happens server-side against the stored artifact, so it does **not** depend on the app being installed, or on the install having succeeded: a chart behind an `installFailed` / `installingCanceled` row is still downloadable.
- **Local path rules mirror `files download`:** omitted → `./<chart>-<version>.tgz` (the name the server suggests); an existing directory (or a path ending in `/`) → that directory with the server-chosen name; anything else → the exact target path.
- **An existing local file is never silently replaced.** Without `--overwrite` the command fails and leaves the file alone. The write goes to `<path>.tmp` and renames on success, so an interrupted transfer cannot truncate a chart you still needed.
- Streams through the no-timeout HTTP client, like `upload`.

## Recovering a chart that only exists on the Olares

```bash
olares-cli market download mychart ./recovered/                     # pull the stored .tgz back
tar -xzf ./recovered/mychart-1.0.0.tgz -C ./recovered/              # unpack to edit
# ... edit the chart, bump metadata.version == Chart.yaml version ...
olares-cli chart package ./recovered/mychart                        # repackage
olares-cli market upload ./mychart-1.0.1.tgz                        # push the new version
olares-cli market upgrade mychart -s upload --version 1.0.1 --watch
```

The new version is not optional: a published version's bytes are immutable, so the edit has to ship under a higher one — see [the immutability rule](olares-market-chart-publish.md#safety-constraints).

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `API error (HTTP 404): Chart not found` | No stored chart for that app in that source | `market list -s upload` to confirm the bucket; pass `-s <id>` for another source |
| `API error (HTTP 501)` | The Olares predates the chart-package endpoint | Upgrade the Olares; there is no client-side fallback |
| `unexpected EOF` partway through | The transfer outlived the market's write timeout — an Olares that predates the longer deadline on this route | Retry on a faster link, or upgrade the Olares; no partial file is left behind |
| `<path> already exists (pass --overwrite to replace it)` | A local file is already at the destination | Pass `--overwrite`, or give a different local path |
| `expected chart bytes but got a JSON response` | The request reached something other than the package route (proxy or version mismatch) | Verify the profile's market URL and the Olares version |

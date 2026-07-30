# doctor `thirdleveldomain`

> Read the parent [`../SKILL.md`](../SKILL.md) first. Flags remain authoritative in `olares-cli doctor thirdleveldomain --help`.

Audits `Application.spec.settings.customDomain.third_level_domain` values per user zone:

```bash
olares-cli doctor thirdleveldomain
olares-cli doctor thirdleveldomain -o json
olares-cli doctor thirdleveldomain --kubeconfig /path/to/config
```

This command uses Kubernetes credentials, not the active Olares profile token. Kubeconfig resolution is `--kubeconfig`, then `KUBECONFIG`, then the default client-go path.

## Findings

- `duplicate`: two or more app/entrance pairs use the same prefix in one user zone.
- `reserved`: the prefix is `auth`, `desktop`, or `wizard`.
- Shared apps are evaluated through each user's effective settings; per-user apps contribute only to their owner's zone.
- Default app-id prefixes and `third_party_domain` are outside this audit.

The read-only audit exits non-zero while findings remain. `-q` suppresses output but preserves that exit-code contract.

## `--force-dedupe` safety

`--force-dedupe` **writes Application CRs**. Before using it:

1. Run the read-only command and review every finding.
2. Obtain explicit user approval for the mutation.
3. Prefer `-o json` when the affected app/entrance set needs machine review.

For duplicate prefixes, it keeps the lexicographically first `(app, entrance)` in each user zone and clears `third_level_domain` on the others. For reserved prefixes, it clears every matching value. Patches retry Kubernetes resource-version conflicts, then the command re-audits and reports any remaining findings.

There is no undo command. Record the original values before applying the fix if restoration may be needed.

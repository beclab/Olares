# settings integration

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings integration accounts --help` and `olares-cli settings integration accounts <verb> --help`.

External integration accounts (S3 / Tencent COS / Google Drive / Dropbox / Olares Space / NFT cloud binding).

## What's covered by the CLI

| Account type | CLI verb | Why |
|---|---|---|
| AWS S3 (or S3-compatible endpoint) | `accounts add awss3` | Direct credentials — no OAuth |
| Tencent COS | `accounts add tencent` | Direct credentials — no OAuth |
| Google Drive | (SPA only) | OAuth-bound browser session |
| Dropbox | (SPA only) | OAuth-bound browser session |
| Olares Space / NFT cloud binding | (SPA only) | Browser- and wallet-bound by design |

> **OAuth and wallet flows stay in the SPA.** The access tokens they produce are scoped to a browser session and have no useful one-shot CLI capture. If the user asks to add Google Drive / Dropbox / Olares Space / NFT integrations via CLI, direct them to the Settings → Integration page in the SPA.

## Sub-tree

| Verb | Floor | Notes |
|---|---|---|
| `accounts list` | normal | `accountMini` shape (no `raw_data`) |
| `accounts list-by-type <type>` | normal | Filter by account type |
| `accounts get <type> [name]` | normal | `accountFull` shape (includes `raw_data`); `name` optional for single-tenant types |
| `accounts add awss3 [flags]` | normal | AWS S3 / S3-compatible |
| `accounts add tencent [flags]` | normal | Tencent COS |
| `accounts delete <type> [name]` | normal | `name` optional for single-tenant types |

## `accounts add awss3`

```bash
olares-cli settings integration accounts add awss3 \
  --access-key-id     "$AWS_ACCESS_KEY_ID" \
  --access-key-secret "$AWS_SECRET_ACCESS_KEY" \
  --endpoint          "https://s3.amazonaws.com" \
  --bucket            "my-bucket"
```

- `--bucket` is **optional** — omit for "any bucket the credentials can reach"; provide for "scope to this bucket".
- `--endpoint` accepts any S3-compatible endpoint (MinIO, Backblaze B2 via S3 API, etc.) — not just AWS.

## `accounts add tencent`

```bash
olares-cli settings integration accounts add tencent \
  --access-key-id     "$TENCENT_SECRET_ID" \
  --access-key-secret "$TENCENT_SECRET_KEY" \
  --endpoint          "https://cos.ap-shanghai.myqcloud.com"
```

- Region is encoded in the endpoint URL — e.g. `cos.ap-beijing.myqcloud.com`, `cos.ap-shanghai.myqcloud.com`.
- Tencent COS is **single-tenant**: no `--bucket` flag, no `<name>` argument on add. There is at most one Tencent account per profile.

## `accounts get` / `accounts delete` — name handling

```bash
# Multi-tenant types (S3, Drive, Dropbox) — need a name.
olares-cli settings integration accounts get awss3 my-bucket
olares-cli settings integration accounts delete awss3 my-bucket

# Single-tenant types (Tencent, Space, NFT) — name omitted.
olares-cli settings integration accounts get tencent
olares-cli settings integration accounts delete tencent
```

The store key is composed as `integration-account:<type>:<name>` (or `integration-account:<type>` when no name is supplied), matching the SPA's `getStoreKey`.

## Secret handling — agent rules

**The single most important rule in this sub-tree: NEVER paste secret-key values into the agent transcript.**

- **Always recommend env vars or stdin pipes**: `--access-key-secret "$AWS_SECRET_ACCESS_KEY"`.
- Bash history retention is the user's responsibility, but the agent's default phrasing should make it easy to keep secrets out of the transcript / scrollback.
- For the agent's own suggestions: write `--access-key-secret "$AWS_SECRET_ACCESS_KEY"` (placeholder), NOT `--access-key-secret "AKIA..."` (real-looking).

## `accounts get` JSON shape

```bash
olares-cli settings integration accounts get awss3 my-bucket -o json
```

Returns the `accountFull` shape including the un-redacted `raw_data` field. **The secret-key value WILL appear in the output** — pipe to `jq` and select only the fields you need rather than dumping the whole payload.

## Agent best practices

- For "add my S3 credentials" → prompt the user to **set the secret in an env var FIRST**, then construct the command using `"$VAR"` interpolation.
- For "show me my integration accounts" → `accounts list` (no secrets) instead of `accounts get` (with secrets) unless the user specifically needs the credential payload.
- For "delete my old S3 account" → `accounts list-by-type awss3` first to confirm the right name, then `accounts delete awss3 <name>`.
- If the user asks for Google Drive / Dropbox / Olares Space / NFT cloud binding, **redirect to the SPA** — explain that OAuth tokens are browser-scoped and can't be captured by one-shot CLI calls.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `account 'X' of type '<type>' not found` | Wrong name / never added | `accounts list-by-type <type>` to enumerate |
| `missing required flag: --access-key-id` (or `--access-key-secret` / `--endpoint`) | Mandatory flag omitted | Provide all three for S3 / Tencent |
| `account already exists` | Single-tenant type (Tencent / Space) already has one | `accounts delete <type>` first, then add |
| Secret value shows up in shell history | The user (or agent) embedded the secret on the command line directly | Re-issue from env var; clear shell history manually |

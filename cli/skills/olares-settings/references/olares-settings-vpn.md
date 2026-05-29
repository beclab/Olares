# settings vpn (Headscale mesh)

> **Prerequisite:** Read [`../../olares-shared/SKILL.md`](../../olares-shared/SKILL.md) and the parent [`../SKILL.md`](../SKILL.md) first.
> **Flags & examples:** `olares-cli settings vpn --help` and `olares-cli settings vpn <noun> <verb> --help`.

Manage the per-Olares Headscale mesh — devices, routes, per-app ACL, SSH toggle, sub-route toggle, public-domain-policy.

## Sub-tree summary

| Sub-tree | Floor | Verbs |
|---|---|---|
| `devices` | normal | `list`, `routes <id>` |
| `ssh` | admin | `status`, `enable`, `disable` |
| `subroutes` | admin | `status` (read-only) |
| `acl` | admin | `all`, `get`, `add`, `remove` (alias `rm`) |
| `public-domain-policy` | admin | `get` (read-only `deny_all` 0/1 flag) |

## `devices` — Headscale machines

```bash
olares-cli settings vpn devices list
olares-cli settings vpn devices routes <device-id>
```

- Hits `<SettingsURL>/headscale/machine` — settings nginx, not desktop.
- **Raw Headscale JSON** (no envelope). `route.id` is a `string`.
- `devices list` is the only normal-floor verb in this sub-tree.

## `ssh` — boolean toggle

```bash
olares-cli settings vpn ssh status
olares-cli settings vpn ssh enable
olares-cli settings vpn ssh disable
```

- `enable` / `disable` send an explicit empty `{}` body to match the SPA's request shape, even though the upstream doesn't read it.
- Run `vpn ssh status` after to confirm the resulting state.

## `subroutes` — read-only toggle

```bash
olares-cli settings vpn subroutes status
```

Returns the `allow_subroutes` flag plus the current sub-route list. `-o json` emits the unwrapped `[]string` the SPA reads.

## `acl` — per-app ACL (RMW under the hood)

The most flag-rich sub-tree here. **The upstream replaces the whole per-app ACL vector on every POST** — there is no add / remove endpoint. `vpn acl add` and `vpn acl remove` are read-modify-write sugar so unrelated entries survive untouched.

### Read

```bash
olares-cli settings vpn acl all                # every app that currently has an ACL row
olares-cli settings vpn acl get my-app         # per-app ACL vector
olares-cli settings vpn acl get my-app -o json # full payload
```

### Add / remove — destination format

Every `--tcp` / `--udp` / `--any-proto` value is a **Headscale destination spec `<host>:<port>`**, NOT a bare port number.

| Spec | Meaning |
|---|---|
| `'*:8080'` | Any source, port 8080 (this is what the Web UI's "Add ACL" implicitly sends) |
| `'192.168.1.0/24:22'` | Restrict source CIDR |
| `'tag:api:443'` | Tag-based source |
| `'example-host:*'` | Allow all ports on one host |
| `'*:8000-8100'` | Port range (Headscale accepts; CLI doesn't validate ranges specifically) |
| `'*:*'` | Allow everything (Headscale accepts) |

**Bare port numbers are rejected** by the Headscale policy parser with `invalid port format`. The CLI now pre-validates the `<host>:<port>` shape and surfaces a copy-pasteable suggestion rather than letting BFL reject the POST.

### `--any-proto` = SPA "Add ACL" parity

The Settings page's "Add ACL" dialog only collects a port, hardcodes the app to the system `olares-app`, sets the source host to `*`, and posts with `proto=""` so Tailscale expands to ICMPv4 / ICMPv6 / TCP / UDP. To mirror that one-for-one:

```bash
# SPA "Add ACL" port 8080 equivalent:
olares-cli settings vpn acl add olares-app --any-proto '*:8080'
```

`--tcp` / `--udp` are strictly more expressive (per-protocol slots, named app, custom hosts/CIDRs/tags) but always emit a typed proto. `add` / `remove` use case-insensitive proto matching on the merge / subtract paths, so empty-proto entries created via `--any-proto` round-trip cleanly alongside `--tcp` / `--udp` ones.

### Examples

```bash
# Merge a single TCP destination.
olares-cli settings vpn acl add my-app --tcp '*:8080'

# Merge multiple destinations (repeated flag OR comma-separated; both work, both de-dupe).
olares-cli settings vpn acl add my-app --tcp '*:80' --tcp '*:443' --udp '*:53'
olares-cli settings vpn acl add my-app --tcp '*:80,*:443'

# Restrict source CIDR.
olares-cli settings vpn acl add my-app --tcp '192.168.1.0/24:22'

# Remove a destination.
olares-cli settings vpn acl remove my-app --tcp '*:80'
olares-cli settings vpn acl rm     my-app --udp '*:53'   # alias

# SPA Web parity (all protocols).
olares-cli settings vpn acl add olares-app --any-proto '*:8080'
```

## `public-domain-policy` — read-only switch

```bash
olares-cli settings vpn public-domain-policy get
```

Returns the `deny_all` flag (0 or 1). The write side stays in the SPA today.

## Agent best practices

- **Always run `vpn acl get <app>` before mutating** — the RMW semantics mean a wrong delta can silently drop unrelated entries (it shouldn't, but inspecting first catches surprise edge cases).
- **For "I want the same thing the SPA's 'Add ACL' button does"** → `vpn acl add olares-app --any-proto '<host>:<port>'`. Don't try to reconstruct it from `--tcp` / `--udp`.
- **Bare port number is the #1 mistake** — when the user says "allow port 8080", suggest `'*:8080'`, not `8080`.
- `ssh enable` / `disable` are SAFE to use in interactive sessions (idempotent boolean toggle), but **note that disabling SSH cuts off any ongoing TermiPass SSH session** — confirm intent before invoking.

## Common errors

| Symptom | Cause | Fix |
|---|---|---|
| `invalid port format` (from Headscale) | Bare port number passed instead of `<host>:<port>` | Use `'*:<port>'` (any source) |
| `acl add: --tcp / --udp / --any-proto required` | No destination specs | Provide at least one |
| `device-id '<id>' not found in Headscale` | Wrong device id | `vpn devices list` to enumerate |
| `role required: ... admin` | Normal user trying admin-only verb | Switch profile to an admin / owner |
| ACL changes don't seem to take effect | Read the RMW caveat — sub-policies / acl entries on other protocols you didn't pass DO survive | Re-run with the right flag set |

# Secrets — user-supplied, middleware-owned, and chart-generated credentials

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md). This reference covers where a credential lives and how it reaches the container. Declaring env vars is [env.md](olares-chart-env.md); middleware wiring is [middleware.md](olares-chart-middleware.md).

A credential reaches a container from a Kubernetes Secret — one key via `secretKeyRef`, a whole set via `envFrom.secretRef`. ConfigMaps, `values.yaml` and literal `env: value:` lines carry configuration; keep credentials in a Secret. This also covers a connection string that embeds a password: it belongs to the Secret, not to the ConfigMap holding the rest of the config.

Which recipe to follow depends on where the value comes from:

| Source of the value | What to write |
|---|---|
| The user, at install | `envs[]` with `type: password` + `required: true`, copied into a chart Secret |
| Olares middleware (Postgres / Redis / MongoDB / …) | omit `password` in the manifest `middleware:` block; reference `.Values.<mw>.password` |
| The chart itself (JWT signing, session, encryption keys) | generate once with `lookup` + `randAlphaNum`, so an upgrade keeps the same value |

## 1. The user supplies it at install

Declare it, then place it in a Secret and reference that Secret from the workload.

```yaml
# OlaresManifest.yaml
envs:
  - envName: ADMIN_PASSWORD
    required: true
    type: password
    editable: false
    regex: '^.{8,}$'
    description: Admin password (min 8 chars)
```

```yaml
# templates/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{ .Release.Name }}-app
  namespace: "{{ .Release.Namespace }}"
type: Opaque
stringData:
  ADMIN_PASSWORD: {{ .Values.olaresEnv.ADMIN_PASSWORD | quote }}
```

```yaml
# templates/deployment.yaml
        env:
        - name: APP_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: {{ .Release.Name }}-app
              key: ADMIN_PASSWORD
```

`stringData` takes the value as-is; `data` expects base64. Use `stringData` when the value arrives from `.Values`, and keep the whole app's credentials in one Secret so the workload can take them with a single `envFrom.secretRef` when there are several.

## 2. Olares middleware owns it

Leave `password` out of the `middleware:` block and Olares manages the credential for that database, handing it to the chart as `.Values.<mw>.password`:

```yaml
# OlaresManifest.yaml
middleware:
  postgres:
    username: myapp
    databases:
    - name: myapp
```

```yaml
# templates/secret.yaml
stringData:
  DB_PASSWORD: {{ .Values.postgres.password | quote }}
  DATABASE_URL: "postgres://{{ .Values.postgres.username }}:{{ .Values.postgres.password }}@{{ .Values.postgres.host }}:{{ .Values.postgres.port }}/{{ .Values.postgres.databases.myapp }}"
```

The same holds for every injected value that is a credential — `.Values.redis.password`, `.Values.mongodb.password`, `.Values.os.appSecret`, `.Values.oidc.client.secret` — and for any config file the app reads: render the file from the Secret, or mount the Secret and let the entrypoint fill the file in.

## 3. The chart generates it

A JWT signing key, a session key or an encryption key has no external source, so the chart mints one. Generate it **once** and reuse it on every later render: a new signing key invalidates every session the app has issued, and a new encryption key leaves the data encrypted with the old one unreadable.

Read the existing Secret first and fall back to a fresh value:

```yaml
{{/* templates/_helpers.tpl */}}
{{- define "myapp.authJwtSecret" -}}
{{- $existing := lookup "v1" "Secret" .Release.Namespace "myapp-auth" -}}
{{- if and $existing $existing.data (index $existing.data "AUTH_JWT_SECRET") -}}
{{- index $existing.data "AUTH_JWT_SECRET" | b64dec -}}
{{- else -}}
{{- randAlphaNum 64 -}}
{{- end -}}
{{- end -}}
```

```yaml
# templates/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-auth
  namespace: "{{ .Release.Namespace }}"
  annotations:
    helm.sh/resource-policy: keep
type: Opaque
data:
  AUTH_JWT_SECRET: {{ include "myapp.authJwtSecret" . | b64enc | quote }}
```

Two mechanics decide whether this works:

- **The `else` branch must produce a value.** `lookup` returns nothing when the chart is rendered without a cluster, which is how `chart lint` renders it. Reading the key unconditionally then fails the render with `index of untyped nil`, and an `if` with no `else` renders the key as an empty string — an app starting with an empty signing key. Write the generator in the `else`.
- **Stay on one side of base64.** `data:` holds base64, so pipe the generated value through `b64enc`. A value read back from `$existing.data` is already base64: `b64dec` it inside the helper, as above, so the helper always returns the plain value and the caller always encodes once. Mixing the two encodes twice, and the app receives a base64 blob where it expects a key.

`helm.sh/resource-policy: keep` keeps the Secret when the release is removed. Add it when data encrypted with the key outlives the release — a database or an appData volume that a later reinstall reattaches. Without it, a reinstall mints a new key and the old data no longer decrypts.

To rotate on purpose, delete the Secret and upgrade: the `else` branch mints a fresh value. Say so in a comment next to the Secret, because it is the one operation that is meant to invalidate sessions.

## Reference charts

Published charts in `beclab/apps` to read for the generate-once pattern: `deerflowv2` (helper plus a documented rotation path), `nemoclaw` (user-pinned value, else stored, else generated), `nofx` (several keys in one Secret), `documenso` (per-key fallback), `wordpress` (app and database Secrets side by side).

## Before packaging

- Every credential the workload consumes comes from a Secret.
- Every chart-generated value goes through `lookup` with a generating `else`.
- The manifest's `middleware:` block declares no `password`.
- Every credential is either supplied at install or generated by the chart, so the chart carries none of its own — which is what a public listing requires ([`../../olares-publish/SKILL.md`](../../olares-publish/SKILL.md)).

# Shared frontend HTTP proxies

Olares mounts the os-frontend Nginx ConfigMaps over the frontend image's bundled
configuration. These routes now follow the user-service proxy changes in TermiPass.

Dashboard/Control Hub BFL and Middleware HTTP, Settings Secret `/admin`, and
Vault BFL and signed `/server` RPC flow through user-service. It verifies the
access token and supplies the verified user and token to the backend. Management
proxies check owner/admin roles; supported self-service routes remain available
to normal users. Monitoring routes connect directly as described below.

The frontend image is `beclab/system-frontend:v1.12.11-alpha.1`; user-service is
`beclab/user-service:v0.1.23-alpha.1`. These routes require the updated
user-service proxy mounts. No image detection or enable flag is introduced.

Dashboard and Control Hub terminal/exec/watch routes connect directly to
`monitoring-server.os-framework:80` (Service port 80 targets container port 8000).
The existing local identity subrequest verifies the session, then Nginx supplies
that verified user and token/Cookie to the backend without a Provider RBAC hop.
HTTP/1.1 upgrades, WebSocket subprotocols, original URI/query and streaming are
preserved; buffering is off and read/send inactivity timeouts are 3600 seconds.
This does not implement an upgrade transport in user-service or assert that
monitoring enforces the removed Provider RBAC policy.

Files and Share transfers retain their browser paths but connect directly to
Files. Login/refresh keep their existing routes. The unused Settings `/vault`
proxy and its legacy upstream map are removed; user-service still calls the
Vault :3010 trust-device API directly, and Vault `/server` RPC is retained.

Headscale's ingress policy explicitly admits `os-frontend` / `app=olares-app` on
8080 and 8000. Port 8001 remains restricted to app-service's policy writer.
The Secret ingress policy admits only that frontend pod to Infisical on 8080;
its labels preserve it through security reconciliation. The general framework,
platform and user-space policies already admit the frontend namespace, which
carries `bytetrade.io/ns-type=system`.

This change does not deploy services or remove system-server. Daemon onboarding,
BFL internal Provider calls and existing users' Helm releases
still need migration. Validate real-cluster networking before deleting providers.

L4 Files node routes now target the Files pod IP on the selected node at port 80.
Node selection stays in L4: node-specific paths go to that node, and generic
master paths go to the master node. `X-Terminus-Node` remains metadata, not a
backend dispatch mechanism; the identity supplied by ExtAuth is preserved.
The Fileserver reconciler watches Files Pod and Node changes, triggering a
Provider snapshot rebuild and xDS refresh when endpoints change. It no longer
creates per-user Services or RBAC, and node discovery does not wait for them.
This does not delete previously created Services or RBAC objects or add another
forwarding layer to the Files backend. File-level permissions and transfers
still need runtime verification before rollout.

All Monitoring HTTP mounts now connect directly to `monitoring-server.os-framework:80`:
Dashboard `/kapis`, `/api`, `/capi`; Control Hub also `/current_user`; Desktop
and Settings `/kapis`. The existing more-specific application routes are retained.
Nginx uses user-service only for the identity subrequest, forwards its verified
user/token, and preserves original paths, query-string watches and stream settings.
The generic user-service Monitoring proxy is no longer used by these frontend
mounts. Its owner/admin check is also bypassed; authorization now depends on the
entry point and Monitoring backend. No backend RBAC equivalence is asserted.

Files frontend fallback routes and Settings `/api/resources`, `/api/nodes/` now
connect directly to `files-service.os-framework:80`, with a local identity
subrequest and verified `X-Bfl-User`/token. Transfers remain in Nginx, preserving
URI rewrites, Range, per-route timeouts and existing 2/4 GB request limits;
response buffering is disabled. Incoming Provider/internal-share routing headers
are cleared. Native direct Seafile paths remain on their original services.

Share fallback routes `/api`, `/upload`, `/videos` connect directly to Files and
retain credentials and Host/forwarded Host. They do not introduce user-service
login checks. Share node routes use the same existing entrance Authelia check as
the share catch-all (including its configured public policy); no new auth bypass
is introduced. Only share resource paths, node uploads and video transfers are
added to L4, not private mount/task/admin paths. Specific node targets precede
master fallbacks. The Files backend retains share-token/password authorization.

The inner Files Nginx 100M body limits are raised to 4000M to match frontend
transfer mounts. This does not change Files/Seafile chunk protocol limits.
Validate anonymous shares, uploads/resume/Range and multi-node transfers in the
cluster before removing system-server. Test deployments do not establish full
protocol or permission equivalence.

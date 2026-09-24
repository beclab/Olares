# Wi-Fi authentication and captive portal integration

## HTTP and BLE contract

The existing connect-wifi HTTP endpoint and BLE connection characteristic accept
the same JSON object. Existing `ssid` and `password` requests remain valid for
unambiguous personal networks. The server resolves authentication from the AP
scan; hidden networks absent from the scan are not supported by this version.

Fields:

- `ssid`: required, 1–32 bytes.
- `password`: personal-network key or 802.1X account password.
- `username`: 802.1X account identity.
- `security`: optional `open`, `owe`, `wep`, `wpa-psk`, `sae`, or `wpa-eap`.
  `wpa-psk` covers WPA/WPA2 Personal. `sae` selects WPA3 Personal.
- `bssid`, `interface`: optional AP and device selectors. Use these when the same
  SSID is advertised with different authentication. An explicit BSSID also pins
  the resulting profile to that AP.
- `wepKeyType`: `key` (default) or `passphrase`; `wepKeyIndex`: 0–3, default 0.
- `enterprise`: optional object described below.

802.1X example using the system trust store:

```json
{
  "ssid": "Office",
  "security": "wpa-eap",
  "username": "alice",
  "password": "account-password",
  "enterprise": { "domainSuffixMatch": "radius.example.com" }
}
```

The default EAP method is PEAP with MSCHAPv2. `enterprise.eap` can be `peap` or
`ttls`. `enterprise.phase2Auth` defaults to `mschapv2`; TTLS also accepts `pap`,
`chap`, and `mschap`. `anonymousIdentity` is optional. EAP-TLS and Suite-B-only
networks are unsupported.

Server certificate validation is enabled by default. Supply `domainSuffixMatch`
and optionally `caCertPath`, an absolute path to an existing CA certificate on
the device. With no CA path the system trust store is used. This API does not
upload certificates. Only explicit `enterprise.insecureSkipVerify: true` disables
server certificate validation; username and password alone do not imply that
validation is disabled.

The scan response includes `ss`, `st`, and `se` (security types array),
`bs` (BSSID), and `if` (interface). Unknown authentication is reported as `unsupported`,
never as an open network. APs are grouped by interface, SSID, and authentication
capabilities; the strongest AP represents a group. HTTP additionally includes
`connected` (always a boolean) and receives the full list. BLE omits connection
status entirely (neither `connected` nor `co`). Its AP list contains only
`ss` (SSID), `st` (signal strength), and `se` (authentication types); `bs` and
`if` are omitted. Equivalent SSID/authentication entries across radios are merged,
keeping the strongest signal. Entries retain
the internal connection/signal priority order; an entry that does not fit is
skipped so a shorter later entry can use the remaining 512-byte capacity. Connection payloads exceeding 512 bytes are
rejected by the BLE parser; no application-level fragmentation is implemented.
Bluetooth stacks may themselves reject an oversized write before it reaches the
server. Use HTTP for larger enterprise requests.

When authentication is omitted, same-name networks with different capabilities
require an explicit selection. On mixed WPA2/WPA3 APs, daemon queries the selected
interface's wpa_supplicant `Capabilities.KeyMgmt` and prefers SAE when supported.
If this capability cannot be queried (including a different Wi-Fi backend),
specify `security` explicitly. Authentication failure never triggers automatic
fallback to weaker security.

Connections have a 90-second deadline, including discovery and activation.
Matching saved profiles retain unrelated settings and their UUID. Credentials are
updated through D-Bus, and profiles are saved after activation succeeds. Failure
restores the previous profile including stored secrets, or deletes a newly
created profile. Errors are English and do not include remote error text or
passwords. A successful response means Wi-Fi activation succeeded, not that
captive portal login has completed.

## Desktop PortalHelper contract

The daemon watches NetworkManager's global `Connectivity` property and
`PrimaryConnection` on the system bus. On PORTAL (2), it opens the primary
connection's login window using `ConnectivityCheckUri`; the desktop browser
handles redirects. It also handles initial PORTAL state, service restarts,
connection switches, and leaving PORTAL. This applies to Ethernet as well as
Wi-Fi. It does not detect a portal on a secondary link while another connection
keeps global connectivity FULL.

Default D-Bus service and interface: `com.olares.desktop.PortalHelper`.
Object: `/com/olares/desktop/PortalHelper`.
See [the interface XML](com.olares.desktop.PortalHelper.xml).

- `OpenPortal(s connectionUuid, s interfaceName, s url)` must enqueue presentation
  and return immediately. Repeated calls for the same connection must not create
  duplicate windows. Accept only HTTP/HTTPS URLs and route the request to the
  local desktop session.
- `ClosePortal(s connectionUuid)` must be idempotent, including when no window is
  open. It must also cancel any queued presentation for that UUID. The service
  should serialize requests for a UUID so an earlier open cannot overtake close.

The daemon uses a five-second call timeout and 1, 2, 4, 8, 16, then 30-second retry
delays. State changes cancel stale calls/retries. A timed-out open is treated as
possibly accepted and closed when the desired connection changes. Successful
opens are deduplicated until the state or service owner changes. Helper failures
do not fail the Wi-Fi connection command. Shutdown closes the tracked window on
a best-effort basis.

`--portal-helper-service=NAME` overrides the bus destination only; the object path
and interface remain fixed. The helper itself is delivered by the desktop, not by
this daemon change.

## Deployment

NetworkManager connectivity checking must already be configured and enabled with
a usable HTTP/HTTPS `ConnectivityCheckUri`. Daemon does not change this setting or
choose a third-party probe endpoint. Invalid/missing URLs are retried and logged
without opening a browser.

The desktop must register the helper on the **system bus** and install its own
service/activation definition and D-Bus policy. For example, with a dedicated
service account `olares-portal` and root olaresd, the policy can grant that account
ownership and root permission to call the two methods:

```xml
<busconfig>
  <policy user="olares-portal">
    <allow own="com.olares.desktop.PortalHelper"/>
  </policy>
  <policy user="root">
    <allow send_destination="com.olares.desktop.PortalHelper"
           send_interface="com.olares.desktop.PortalHelper"
           send_member="OpenPortal"/>
    <allow send_destination="com.olares.desktop.PortalHelper"
           send_interface="com.olares.desktop.PortalHelper"
           send_member="ClosePortal"/>
  </policy>
</busconfig>
```

Adapt the service account to the actual desktop deployment; do not install this
example verbatim without that account. Keep ownership restricted to the helper,
and validate sender credentials in the helper. The desktop owns session routing
and the sandboxed authentication UI. Service-name overrides require matching
policy changes. Portal window dismissal/login may trigger NetworkManager's
`CheckConnectivity` from the desktop to accelerate completion detection.

## Validation

```sh
go test -race ./internel/wifi ./internel/portal
go test ./internel/ble ./internel/apiserver/handlers -run 'TestWifi|TestConnectWifi'
go build -o /tmp/olaresd ./cmd/terminusd
```

The Portal test uses an isolated real D-Bus daemon if `dbus-daemon` is installed;
it skips that test otherwise. Other tests mock NetworkManager operations and
exercise authentication mappings, profile rollback, timeouts, retries, and
payload limits. The pre-existing BLE `TestScan` requires physical hardware and is
not part of these commands.

Hardware acceptance: connect to open/OWE, WPA2/WPA3, PEAP/MSCHAPv2, and TTLS APs;
verify invalid credentials fail without erasing saved settings; exercise a real
portal redirect and login; switch primary networks during login and restart NM
and the helper. Confirm only one current login window remains and credentials do
not appear in logs. These checks require a Linux device with Wi-Fi and the desktop
helper implementation.

References: [NM D-Bus types](https://networkmanager.dev/docs/api/latest/nm-dbus-types.html),
[NM properties](https://networkmanager.dev/docs/api/latest/gdbus-org.freedesktop.NetworkManager.html),
[802.1X settings](https://networkmanager.dev/docs/api/latest/settings-802-1x.html),
[wpa_supplicant D-Bus](https://w1.fi/wpa_supplicant/devel/dbus.html).

## Standalone AP list endpoint

`GET /system/list-aps` uses the same `RequireLocal` middleware as `/system/ifs`.
That middleware currently passes requests through without enforcing access control.
It scans NetworkManager directly, independently of Bluetooth initialization or
its cached AP list. Like BLE scanning, it enables a disabled Wi-Fi radio before
scanning. The scan has a 30-second deadline and returns the full list, ordered by
connected status and signal strength. Existing grouping by interface, SSID and
security capabilities applies.

```json
{
  "code": 200,
  "message": "",
  "data": [
    { "ss": "Office", "st": 85, "connected": true,
      "se": ["wpa-eap"], "bs": "00:11:22:33:44:55", "if": "wlan0" }
  ]
}
```

No Wi-Fi devices or no visible APs returns `data: []`. NetworkManager failures
or scan timeouts return HTTP 503 with an English error message. This endpoint
does not apply the BLE 512-byte limit and does not return credentials.

package meshinagent

import (
	"fmt"
	"regexp"
	"strings"
)

// BearerJS is the njs module for JWT reads, Host allowlist checks, and stream
// SNI decideOffload (allowlist → loopback terminate; else host:443 passthrough).
func BearerJS() string {
	return BearerJSWith(JWTSecretMountPath+"/token", HostsMountPath+"/"+SharedHostsFileName, HTTPSTerminatePort, "olares.com")
}

// BearerJSWith builds the njs module with explicit paths (tests / render).
func BearerJSWith(jwtPath, hostsFile string, terminatePort int, platformDomain string) string {
	escapedDomain := regexp.QuoteMeta(strings.ToLower(strings.TrimSpace(platformDomain)))
	if escapedDomain == "" {
		escapedDomain = "olares\\.com"
	}
	return fmt.Sprintf(`var fs = require('fs');

const JWT_PATH = '%s';
const HOSTS_FILE = '%s';
const TERMINATE = '127.0.0.1:%d';
const CACHE_TTL_MS = 5000;
const PLATFORM_SUFFIX = '.%s';
const V2_GUARD = new RegExp('^[a-z0-9-]+\\.shared\\.%s$', 'i');

let cachedHosts = null;
let cachedMtimeMs = 0;
let cachedAtMs = 0;

function nowMs() { return Date.now(); }

function normalizeHost(v) {
  if (!v) { return ''; }
  return ('' + v).trim().toLowerCase();
}

function passthrough(host) {
  if (!host) { return 'invalid.local:443'; }
  return host + ':443';
}

function isPlatformHost(host) {
  return host.length > PLATFORM_SUFFIX.length && host.endsWith(PLATFORM_SUFFIX);
}

function parseHosts(content) {
  const out = {};
  const lines = content.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim().toLowerCase();
    if (!line || line.startsWith('#')) { continue; }
    out[line] = true;
  }
  return out;
}

function reloadHostsIfNeeded() {
  const now = nowMs();
  if (cachedHosts !== null && now - cachedAtMs < CACHE_TTL_MS) {
    return cachedHosts;
  }
  try {
    const st = fs.statSync(HOSTS_FILE);
    if (cachedHosts !== null && st.mtimeMs === cachedMtimeMs && now - cachedAtMs < CACHE_TTL_MS) {
      return cachedHosts;
    }
    const content = fs.readFileSync(HOSTS_FILE, 'utf8');
    cachedHosts = parseHosts(content);
    cachedMtimeMs = st.mtimeMs;
    cachedAtMs = now;
    return cachedHosts;
  } catch (e) {
    cachedAtMs = now;
    if (cachedHosts === null) { cachedHosts = {}; }
    return cachedHosts;
  }
}

function readJWT(r) {
  try {
    var t = fs.readFileSync(JWT_PATH);
    if (t === undefined || t === null) { return ''; }
    return t.toString().replace(/\s+$/g, '');
  } catch (e) {
    return '';
  }
}

function checkHost(r) {
  const host = normalizeHost(r.headersIn.host || r.variables.host || '');
  if (!host) { return '0'; }
  if (V2_GUARD.test(host)) { return '0'; }
  const hosts = reloadHostsIfNeeded();
  if (hosts[host]) { return '1'; }
  return '0';
}

function decideOffload(s) {
  const host = normalizeHost(s.variables.ssl_preread_server_name);
  if (!host) { return passthrough(host); }
  if (V2_GUARD.test(host)) { return passthrough(host); }
  if (!isPlatformHost(host)) { return passthrough(host); }
  const hosts = reloadHostsIfNeeded();
  if (hosts[host]) { return TERMINATE; }
  return passthrough(host);
}

export default {readJWT, checkHost, decideOffload};
`, jwtPath, hostsFile, terminatePort, escapedDomain, escapedDomain)
}

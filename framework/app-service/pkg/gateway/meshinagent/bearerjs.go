package meshinagent

import (
	"fmt"
	"regexp"
	"strings"
)

// BearerJS is the njs module for JWT reads, auth/tls host allowlists, and stream
// SNI decideOffload (tls-hosts → loopback terminate; else host:443 passthrough).
func BearerJS() string {
	return BearerJSWith(
		JWTSecretMountPath+"/token",
		HostsMountPath+"/"+SharedHostsFileName,
		HostsMountPath+"/"+TLSHostsFileName,
		HTTPSTerminatePort,
		"olares.com",
		CertsMountPath,
		PlaceholderCertDir,
	)
}

// BearerJSWith builds the njs module with explicit paths (tests / render).
func BearerJSWith(jwtPath, authHostsFile, tlsHostsFile string, terminatePort int, platformDomain, certDir, placeholderDir string) string {
	escapedDomain := regexp.QuoteMeta(strings.ToLower(strings.TrimSpace(platformDomain)))
	if escapedDomain == "" {
		escapedDomain = "olares\\.com"
	}
	if certDir == "" {
		certDir = CertsMountPath
	}
	if placeholderDir == "" {
		placeholderDir = PlaceholderCertDir
	}
	return fmt.Sprintf(`var fs = require('fs');

const JWT_PATH = '%s';
const AUTH_HOSTS_FILE = '%s';
const TLS_HOSTS_FILE = '%s';
const TERMINATE = '127.0.0.1:%d';
const CACHE_TTL_MS = 5000;
const PLATFORM_SUFFIX = '.%s';
const REAL_CERT = '%s/tls.crt';
const REAL_KEY = '%s/tls.key';
const PH_CERT = '%s/tls.crt';
const PH_KEY = '%s/tls.key';

let cachedAuthHosts = null;
let cachedAuthMtimeMs = 0;
let cachedAuthAtMs = 0;
let cachedTLSHosts = null;
let cachedTLSMtimeMs = 0;
let cachedTLSAtMs = 0;
let tlsModeCached = null;
let tlsModeAtMs = 0;

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

function reloadHostsFile(path, cached, mtimeMs, atMs) {
  const now = nowMs();
  if (cached !== null && now - atMs < CACHE_TTL_MS) {
    return { hosts: cached, mtimeMs: mtimeMs, atMs: atMs };
  }
  try {
    const st = fs.statSync(path);
    if (cached !== null && st.mtimeMs === mtimeMs && now - atMs < CACHE_TTL_MS) {
      return { hosts: cached, mtimeMs: mtimeMs, atMs: atMs };
    }
    const content = fs.readFileSync(path, 'utf8');
    return { hosts: parseHosts(content), mtimeMs: st.mtimeMs, atMs: now };
  } catch (e) {
    return { hosts: cached === null ? {} : cached, mtimeMs: mtimeMs, atMs: now };
  }
}

function reloadAuthHosts() {
  const r = reloadHostsFile(AUTH_HOSTS_FILE, cachedAuthHosts, cachedAuthMtimeMs, cachedAuthAtMs);
  cachedAuthHosts = r.hosts;
  cachedAuthMtimeMs = r.mtimeMs;
  cachedAuthAtMs = r.atMs;
  return cachedAuthHosts;
}

function reloadTLSHosts() {
  const r = reloadHostsFile(TLS_HOSTS_FILE, cachedTLSHosts, cachedTLSMtimeMs, cachedTLSAtMs);
  cachedTLSHosts = r.hosts;
  cachedTLSMtimeMs = r.mtimeMs;
  cachedTLSAtMs = r.atMs;
  return cachedTLSHosts;
}

function realCertReady() {
  try {
    const c = fs.statSync(REAL_CERT);
    const k = fs.statSync(REAL_KEY);
    return c.size > 0 && k.size > 0;
  } catch (e) {
    return false;
  }
}

function tlsMode(r) {
  const now = nowMs();
  if (tlsModeCached !== null && now - tlsModeAtMs < CACHE_TTL_MS) {
    return tlsModeCached;
  }
  tlsModeCached = realCertReady() ? 'real' : 'placeholder';
  tlsModeAtMs = now;
  return tlsModeCached;
}

function pickCert(r) {
  return tlsMode(r) === 'real' ? REAL_CERT : PH_CERT;
}

function pickKey(r) {
  return tlsMode(r) === 'real' ? REAL_KEY : PH_KEY;
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

function jwtEligible(host) {
  if (!host) { return false; }
  const hosts = reloadAuthHosts();
  return !!hosts[host];
}

function checkHost(r) {
  const host = normalizeHost(r.headersIn.host || r.variables.host || '');
  if (!host) { return '0'; }
  const hosts = reloadTLSHosts();
  if (hosts[host]) { return '1'; }
  return '0';
}

function callerJwt(r) {
  const host = normalizeHost(r.headersIn.host || r.variables.host || '');
  if (!jwtEligible(host)) { return ''; }
  return readJWT(r);
}

function authDeny(r) {
  const host = normalizeHost(r.headersIn.host || r.variables.host || '');
  if (!jwtEligible(host)) { return '0'; }
  if (!readJWT(r)) { return '1'; }
  return '0';
}

function decideOffload(s) {
  // Empty SNI means the connection was redirected by IP (no hostname). Under
  // REDIRECT the original destination is unrecoverable — log and fail closed.
  const host = normalizeHost(s.variables.ssl_preread_server_name);
  if (!host) {
    s.error('mesh-in: no SNI on hijacked 443; original destination unrecoverable under REDIRECT, connection dropped');
    return passthrough(host);
  }
  if (!isPlatformHost(host)) { return passthrough(host); }
  const hosts = reloadTLSHosts();
  if (hosts[host]) { return TERMINATE; }
  return passthrough(host);
}

export default {readJWT, checkHost, callerJwt, authDeny, decideOffload, tlsMode, pickCert, pickKey};
`, jwtPath, authHostsFile, tlsHostsFile, terminatePort, escapedDomain, certDir, certDir, placeholderDir, placeholderDir)
}

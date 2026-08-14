package meshinagent

import (
	"fmt"
)

// BearerJS is the njs module for JWT reads, auth/tls host allowlists, and stream
// SNI decideOffload (tls-hosts → loopback terminate; else host:443 passthrough).
func BearerJS() string {
	return BearerJSWith(
		JWTSecretMountPath+"/token",
		HostsMountPath+"/"+SharedHostsFileName,
		HostsMountPath+"/"+TLSHostsFileName,
		HostsMountPath+"/"+CustomTLSHostsFileName,
		HTTPSTerminatePort,
		CertsMountPath,
		PlaceholderCertDir,
		CustomCertsMountPath,
	)
}

// BearerJSWith builds the njs module with explicit paths (tests / render).
func BearerJSWith(jwtPath, authHostsFile, tlsHostsFile, customTLSHostsFile string, terminatePort int, certDir, placeholderDir, customCertsDir string) string {
	if certDir == "" {
		certDir = CertsMountPath
	}
	if placeholderDir == "" {
		placeholderDir = PlaceholderCertDir
	}
	if customCertsDir == "" {
		customCertsDir = CustomCertsMountPath
	}
	if customTLSHostsFile == "" {
		customTLSHostsFile = HostsMountPath + "/" + CustomTLSHostsFileName
	}
	return fmt.Sprintf(`var fs = require('fs');

const JWT_PATH = '%s';
const AUTH_HOSTS_FILE = '%s';
const TLS_HOSTS_FILE = '%s';
const CUSTOM_TLS_HOSTS_FILE = '%s';
const TERMINATE = '127.0.0.1:%d';
const CACHE_TTL_MS = 5000;
const REAL_CERT = '%s/tls.crt';
const REAL_KEY = '%s/tls.key';
const PH_CERT = '%s/tls.crt';
const PH_KEY = '%s/tls.key';
const CUSTOM_CERTS_DIR = '%s';
// Nonexistent paths: nginx aborts the handshake instead of presenting a wrong cert.
const REJECT_CERT = CUSTOM_CERTS_DIR + '/.reject/tls.crt';
const REJECT_KEY = CUSTOM_CERTS_DIR + '/.reject/tls.key';

let cachedAuthHosts = null;
let cachedAuthMtimeMs = 0;
let cachedAuthAtMs = 0;
let cachedTLSHosts = null;
let cachedTLSMtimeMs = 0;
let cachedTLSAtMs = 0;
let cachedCustomTLSHosts = null;
let cachedCustomTLSMtimeMs = 0;
let cachedCustomTLSAtMs = 0;
let tlsModeCached = null;
let tlsModeAtMs = 0;
let tlsModeCachedHost = '';

function nowMs() { return Date.now(); }

function normalizeHost(v) {
  if (!v) { return ''; }
  return ('' + v).trim().toLowerCase();
}

function passthrough(host) {
  if (!host) { return 'invalid.local:443'; }
  return host + ':443';
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

function reloadCustomTLSHosts() {
  const r = reloadHostsFile(CUSTOM_TLS_HOSTS_FILE, cachedCustomTLSHosts, cachedCustomTLSMtimeMs, cachedCustomTLSAtMs);
  cachedCustomTLSHosts = r.hosts;
  cachedCustomTLSMtimeMs = r.mtimeMs;
  cachedCustomTLSAtMs = r.atMs;
  return cachedCustomTLSHosts;
}

function customTLSRequired(host) {
  if (!host) { return false; }
  return !!reloadCustomTLSHosts()[host];
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

function requestHost(r) {
  if (!r) { return ''; }
  if (r.headersIn && r.headersIn.host) {
    return normalizeHost(r.headersIn.host);
  }
  if (r.variables) {
    return normalizeHost(r.variables.ssl_server_name || r.variables.host || '');
  }
  return '';
}

function resolveCertPair(host) {
  if (!host) { return null; }
  const cert = CUSTOM_CERTS_DIR + '/' + host + '.crt';
  const key = CUSTOM_CERTS_DIR + '/' + host + '.key';
  try {
    if (fs.statSync(cert).size > 0 && fs.statSync(key).size > 0) {
      return { cert: cert, key: key };
    }
  } catch (e) {}
  return null;
}

function customCertReady(host) {
  return resolveCertPair(host) !== null;
}

function tlsMode(r) {
  const host = requestHost(r);
  const now = nowMs();
  if (tlsModeCached !== null && tlsModeCachedHost === host && now - tlsModeAtMs < CACHE_TTL_MS) {
    return tlsModeCached;
  }
  // Common path: platform Hosts are not on custom-tls-hosts — skip custom stats.
  if (!customTLSRequired(host)) {
    tlsModeCached = realCertReady() ? 'real' : 'placeholder';
  } else {
    tlsModeCached = customCertReady(host) ? 'real' : 'reject';
  }
  tlsModeCachedHost = host;
  tlsModeAtMs = now;
  return tlsModeCached;
}

function pickCert(r) {
  const host = requestHost(r);
  // Third-party FQDNs are rare: gate on custom-tls-hosts before any disk stat.
  if (customTLSRequired(host)) {
    const pair = resolveCertPair(host);
    if (pair) { return pair.cert; }
    return REJECT_CERT;
  }
  return tlsMode(r) === 'real' ? REAL_CERT : PH_CERT;
}

function pickKey(r) {
  const host = requestHost(r);
  if (customTLSRequired(host)) {
    const pair = resolveCertPair(host);
    if (pair) { return pair.key; }
    return REJECT_KEY;
  }
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
  // Offload is allowlist-only (tls-hosts): no hard-coded platform suffix, so
  // any zone the control plane materialized (e.g. .olares.cn) can terminate.
  const host = normalizeHost(s.variables.ssl_preread_server_name);
  if (!host) {
    s.error('mesh-in: no SNI on hijacked 443; original destination unrecoverable under REDIRECT, connection dropped');
    return passthrough(host);
  }
  const hosts = reloadTLSHosts();
  if (hosts[host]) { return TERMINATE; }
  return passthrough(host);
}

export default {readJWT, checkHost, callerJwt, authDeny, decideOffload, tlsMode, pickCert, pickKey};
`, jwtPath, authHostsFile, tlsHostsFile, customTLSHostsFile, terminatePort, certDir, certDir, placeholderDir, placeholderDir, customCertsDir)
}

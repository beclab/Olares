#!/usr/bin/env bash
# Re-sign Linkerd identity issuer using existing trust anchor (ca.crt + ca.key).
# Requires ca.crt and ca.key in the same directory; overwrites issuer.crt and issuer.key.
set -euo pipefail

DIR="${1:-}"
if [[ -z "${DIR}" ]]; then
  echo "usage: $0 <dir-with-ca.key-and-ca.crt>" >&2
  exit 1
fi

ISSUER_DAYS="${OLARES_LINKERD_ISSUER_DAYS:-1095}"

CA_CRT="${DIR}/ca.crt"
CA_KEY="${DIR}/ca.key"
ISSUER_KEY="${DIR}/issuer.key"
ISSUER_CSR="${DIR}/issuer.csr"
ISSUER_CRT="${DIR}/issuer.crt"
ISSUER_EXT="${DIR}/issuer-ext.cnf"

[[ -f "${CA_CRT}" && -f "${CA_KEY}" ]] || {
  echo "missing ${CA_CRT} or ${CA_KEY}" >&2
  exit 1
}

openssl ecparam -name prime256v1 -genkey -noout -out "${ISSUER_KEY}"
openssl req -new -key "${ISSUER_KEY}" -out "${ISSUER_CSR}" \
  -subj "/CN=identity.linkerd.cluster.local"

cat > "${ISSUER_EXT}" <<'EOF'
[ v3_intermediate_ca ]
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, keyCertSign, cRLSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
EOF

openssl x509 -req -in "${ISSUER_CSR}" -CA "${CA_CRT}" -CAkey "${CA_KEY}" -CAcreateserial \
  -out "${ISSUER_CRT}" -days "${ISSUER_DAYS}" -sha256 \
  -extfile "${ISSUER_EXT}" -extensions v3_intermediate_ca

chmod 600 "${ISSUER_KEY}" 2>/dev/null || true
rm -f "${ISSUER_CSR}" "${ISSUER_EXT}" "${DIR}/ca.srl"

echo "OK: rotated issuer in ${DIR} (${ISSUER_DAYS}d)"

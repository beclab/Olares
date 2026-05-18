#!/usr/bin/env bash
# Generate Linkerd trust anchor (30y) + intermediate issuer (3y), ECDSA P-256.
# See https://linkerd.io/2/tasks/generate-certificates/
set -euo pipefail

OUT_DIR="${1:-}"
if [[ -z "${OUT_DIR}" ]]; then
  echo "usage: $0 <output-dir>" >&2
  exit 1
fi

CA_DAYS="${OLARES_LINKERD_CA_DAYS:-10950}"
ISSUER_DAYS="${OLARES_LINKERD_ISSUER_DAYS:-1095}"

mkdir -p "${OUT_DIR}"

CA_KEY="${OUT_DIR}/ca.key"
CA_CRT="${OUT_DIR}/ca.crt"
ISSUER_KEY="${OUT_DIR}/issuer.key"
ISSUER_CSR="${OUT_DIR}/issuer.csr"
ISSUER_CRT="${OUT_DIR}/issuer.crt"
ISSUER_EXT="${OUT_DIR}/issuer-ext.cnf"

openssl ecparam -name prime256v1 -genkey -noout -out "${CA_KEY}"
openssl req -x509 -new -key "${CA_KEY}" -sha256 -days "${CA_DAYS}" -out "${CA_CRT}" \
  -subj "/CN=root.linkerd.cluster.local" \
  -addext "basicConstraints=critical,CA:true" \
  -addext "keyUsage=critical,keyCertSign,cRLSign"

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

chmod 600 "${CA_KEY}" "${ISSUER_KEY}" 2>/dev/null || true
rm -f "${ISSUER_CSR}" "${ISSUER_EXT}" "${OUT_DIR}/ca.srl"

echo "OK: wrote ${CA_CRT} (${CA_DAYS}d) ${ISSUER_CRT} (${ISSUER_DAYS}d) ${ISSUER_KEY}"

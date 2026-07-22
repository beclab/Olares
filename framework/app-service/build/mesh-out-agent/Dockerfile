# syntax=docker/dockerfile:1.7
#
# Product image: beclab/mesh-out-agent (engine: nginx alpine-slim).
# Product name: beclab/mesh-out-agent:<ver>-<rev> (REF-SH-AGENT-STACK-01 §2).

ARG NGINX_ALPINE_SLIM_TAG=1.30.0-alpine-slim

FROM nginx:${NGINX_ALPINE_SLIM_TAG}

ARG NGINX_VERSION=1.30.0
ARG MESH_OUT_OLARES_REV=dev
ARG SOURCE_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN set -eux; \
    nginx -V 2>&1 | tr ' ' '\n' | grep -E -- '--with-http_ssl_module$' >/dev/null \
        || (echo "FATAL: --with-http_ssl_module missing in alpine-slim base" >&2; exit 1); \
    test ! -f /etc/nginx/modules/ngx_stream_js_module.so \
        || (echo "FATAL: unexpected njs stream module in slim image" >&2; exit 1); \
    test ! -f /etc/nginx/modules/ngx_http_js_module.so \
        || (echo "FATAL: unexpected njs http module in slim image" >&2; exit 1)

LABEL org.opencontainers.image.title="beclab/mesh-out-agent" \
      org.opencontainers.image.description="Olares mesh-out-agent (engine: nginx ${NGINX_VERSION} alpine-slim)" \
      org.opencontainers.image.vendor="beclab / Olares" \
      org.opencontainers.image.source="https://github.com/beclab/Olares" \
      org.opencontainers.image.version="${NGINX_VERSION}-${MESH_OUT_OLARES_REV}" \
      org.opencontainers.image.revision="${SOURCE_COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.licenses="BSD-2-Clause" \
      io.olares.image.role="mesh-out-agent" \
      io.olares.image.engine="nginx-alpine-slim" \
      io.olares.image.wi="WI-OC-L2B-01"

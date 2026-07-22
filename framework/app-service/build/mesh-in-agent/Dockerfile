# syntax=docker/dockerfile:1.7
#
# Product image: beclab/mesh-in-agent (engine: nginx alpine + njs).
# Product name: beclab/mesh-in-agent:<ver>-<rev> (REF-SH-AGENT-STACK-01 §2).

ARG NGINX_ALPINE_TAG=1.30.0-alpine

FROM nginx:${NGINX_ALPINE_TAG}

ARG NGINX_VERSION=1.30.0
ARG NJS_OLARES_REV=dev
ARG SOURCE_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN set -eux; \
    apk add --no-cache nginx-module-njs; \
    nginx -V 2>&1 | tr ' ' '\n' | grep -E -- '--with-stream(_ssl(_module|_preread_module))?$|--with-http_ssl_module|--with-http_v2_module' >/dev/null \
        || (echo "FATAL: required nginx modules missing in base image" >&2; exit 1); \
    test -f /etc/nginx/modules/ngx_stream_js_module.so \
        || (echo "FATAL: ngx_stream_js_module.so not installed by nginx-module-njs" >&2; exit 1); \
    test -f /etc/nginx/modules/ngx_http_js_module.so \
        || (echo "FATAL: ngx_http_js_module.so not installed by nginx-module-njs" >&2; exit 1)

LABEL org.opencontainers.image.title="beclab/mesh-in-agent" \
      org.opencontainers.image.description="Olares mesh-in-agent (engine: nginx ${NGINX_VERSION} alpine + njs)" \
      org.opencontainers.image.vendor="beclab / Olares" \
      org.opencontainers.image.source="https://github.com/beclab/Olares" \
      org.opencontainers.image.version="${NGINX_VERSION}-${NJS_OLARES_REV}" \
      org.opencontainers.image.revision="${SOURCE_COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.licenses="BSD-2-Clause" \
      io.olares.image.role="mesh-in-agent" \
      io.olares.image.engine="nginx-njs" \
      io.olares.image.wi="WI-OC-L1C-01"

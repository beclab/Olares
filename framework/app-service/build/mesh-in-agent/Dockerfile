# syntax=docker/dockerfile:1.7
#
# beclab/mesh-in-agent — nginx alpine + njs; process UID 1651.
# Build context: framework/app-service/build/mesh-in-agent

ARG NGINX_ALPINE_TAG=1.30.0-alpine
ARG MESH_IN_UID=1651
ARG GO_VERSION=1.24.3

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS ensure-placeholder-cert
ARG TARGETARCH
WORKDIR /src
COPY ensure-placeholder-cert/go.mod ensure-placeholder-cert/main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" \
      -o /out/ensure-placeholder-cert .

FROM nginx:${NGINX_ALPINE_TAG}

ARG NGINX_ALPINE_TAG=1.30.0-alpine
ARG NGINX_VERSION=1.30.0
ARG NJS_OLARES_REV=r5
ARG SOURCE_COMMIT=unknown
ARG BUILD_DATE=unknown
ARG MESH_IN_UID=1651

COPY --from=ensure-placeholder-cert /out/ensure-placeholder-cert /usr/local/bin/ensure-placeholder-cert

RUN set -eux; \
    apk add --no-cache nginx-module-njs iptables bind-tools; \
    NGINX_V="$(nginx -V 2>&1)"; \
    for flag in \
        --with-stream \
        --with-stream_ssl_module \
        --with-stream_ssl_preread_module \
        --with-http_ssl_module \
        --with-http_v2_module; do \
      echo "$NGINX_V" | tr ' ' '\n' | grep -qx -- "$flag" \
        || (echo "FATAL: nginx -V missing required flag: $flag" >&2; exit 1); \
    done; \
    test -f /etc/nginx/modules/ngx_stream_js_module.so \
        || (echo "FATAL: ngx_stream_js_module.so not installed by nginx-module-njs" >&2; exit 1); \
    test -f /etc/nginx/modules/ngx_http_js_module.so \
        || (echo "FATAL: ngx_http_js_module.so not installed by nginx-module-njs" >&2; exit 1); \
    command -v iptables >/dev/null \
        || (echo "FATAL: iptables missing (required by mesh-in iptables init)" >&2; exit 1); \
    test -x /usr/local/bin/ensure-placeholder-cert \
        || (echo "FATAL: ensure-placeholder-cert missing" >&2; exit 1); \
    addgroup -g "${MESH_IN_UID}" -S meshin; \
    adduser -u "${MESH_IN_UID}" -S -G meshin -H -D -h /var/cache/nginx -s /sbin/nologin meshin; \
    mkdir -p /tmp/mesh-in /tmp/olares/mesh-in-placeholder /var/log/nginx /var/cache/nginx /var/run; \
    chown -R meshin:meshin /var/cache/nginx /var/log/nginx /tmp/mesh-in /tmp/olares; \
    if grep -qE '^[[:space:]]*user[[:space:]]+' /etc/nginx/nginx.conf; then \
      sed -i "s/^[[:space:]]*user[[:space:]].*/user  meshin;/" /etc/nginx/nginx.conf; \
    fi; \
    NGINX_VER="$(nginx -v 2>&1 | sed -n 's/^nginx version: nginx\///p')"; \
    NJS_APK="$(apk info -e -v nginx-module-njs | head -n1)"; \
    NJS_VER="$(printf '%s\n' "$NJS_APK" | sed -n 's/^nginx-module-njs-[0-9.]*\.0\.\([0-9.]*\)-r[0-9]*$/\1/p')"; \
    if [ -z "$NJS_VER" ]; then NJS_VER="$(printf '%s\n' "$NJS_APK" | sed -n 's/^nginx-module-njs-//p')"; fi; \
    mkdir -p /usr/share/doc/mesh-in-agent; \
    { \
      echo "base_image=nginx:${NGINX_ALPINE_TAG}"; \
      echo "nginx_version=${NGINX_VER}"; \
      echo "mesh_in_uid=${MESH_IN_UID}"; \
      echo "nginx_module_njs_apk=${NJS_APK}"; \
      echo "njs_version=${NJS_VER}"; \
      echo "extra_apk=iptables,bind-tools"; \
      echo "helper=ensure-placeholder-cert"; \
    } > /usr/share/doc/mesh-in-agent/UPSTREAM_SOURCES.txt

USER root

LABEL org.opencontainers.image.title="beclab/mesh-in-agent" \
      org.opencontainers.image.description="Olares mesh-in-agent (nginx ${NGINX_VERSION} alpine + njs)" \
      org.opencontainers.image.vendor="beclab / Olares" \
      org.opencontainers.image.source="https://github.com/beclab/Olares" \
      org.opencontainers.image.version="${NGINX_VERSION}-${NJS_OLARES_REV}" \
      org.opencontainers.image.revision="${SOURCE_COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.licenses="BSD-2-Clause" \
      io.olares.image.role="mesh-in-agent" \
      io.olares.image.engine="nginx-njs" \
      io.olares.image.uid="${MESH_IN_UID}"

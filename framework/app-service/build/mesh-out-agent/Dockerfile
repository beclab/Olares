# syntax=docker/dockerfile:1.7
ARG NGINX_ALPINE_SLIM_TAG=1.30.0-alpine-slim
FROM nginx:${NGINX_ALPINE_SLIM_TAG}

ARG NGINX_ALPINE_SLIM_TAG=1.30.0-alpine-slim

RUN set -eux; \
    nginx -V 2>&1 | tr ' ' '\n' | grep -E -- '--with-http_ssl_module$' >/dev/null \
        || (echo "FATAL: --with-http_ssl_module missing in alpine-slim base" >&2; exit 1); \
    test ! -f /etc/nginx/modules/ngx_stream_js_module.so; \
    test ! -f /etc/nginx/modules/ngx_http_js_module.so; \
    NGINX_VER="$(nginx -v 2>&1 | sed -n 's/^nginx version: nginx\///p')"; \
    mkdir -p /usr/share/doc/mesh-out-agent; \
    { \
      echo "base_image=nginx:${NGINX_ALPINE_SLIM_TAG}"; \
      echo "nginx_version=${NGINX_VER}"; \
      echo "nginx_git=https://github.com/nginx/nginx"; \
      echo "nginx_checkout=git clone https://github.com/nginx/nginx.git && git -C nginx checkout release-${NGINX_VER}"; \
      echo "nginx_tarball=https://nginx.org/download/nginx-${NGINX_VER}.tar.gz"; \
    } > /usr/share/doc/mesh-out-agent/UPSTREAM_SOURCES.txt

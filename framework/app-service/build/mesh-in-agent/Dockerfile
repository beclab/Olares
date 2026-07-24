# syntax=docker/dockerfile:1.7
ARG NGINX_ALPINE_TAG=1.30.0-alpine
FROM nginx:${NGINX_ALPINE_TAG}

ARG NGINX_ALPINE_TAG=1.30.0-alpine

RUN set -eux; \
    apk add --no-cache nginx-module-njs; \
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
    test -f /etc/nginx/modules/ngx_stream_js_module.so; \
    test -f /etc/nginx/modules/ngx_http_js_module.so; \
    NGINX_VER="$(nginx -v 2>&1 | sed -n 's/^nginx version: nginx\///p')"; \
    NJS_APK="$(apk info -e -v nginx-module-njs | head -n1)"; \
    NJS_VER="$(printf '%s\n' "$NJS_APK" | sed -n 's/^nginx-module-njs-[0-9.]*\.0\.\([0-9.]*\)-r[0-9]*$/\1/p')"; \
    if [ -z "$NJS_VER" ]; then NJS_VER="$(printf '%s\n' "$NJS_APK" | sed -n 's/^nginx-module-njs-//p')"; fi; \
    mkdir -p /usr/share/doc/mesh-in-agent; \
    { \
      echo "base_image=nginx:${NGINX_ALPINE_TAG}"; \
      echo "nginx_version=${NGINX_VER}"; \
      echo "nginx_git=https://github.com/nginx/nginx"; \
      echo "nginx_checkout=git clone https://github.com/nginx/nginx.git && git -C nginx checkout release-${NGINX_VER}"; \
      echo "nginx_tarball=https://nginx.org/download/nginx-${NGINX_VER}.tar.gz"; \
      echo "nginx_module_njs_apk=${NJS_APK}"; \
      echo "njs_version=${NJS_VER}"; \
      echo "njs_git=https://github.com/nginx/njs"; \
      echo "njs_checkout=git clone https://github.com/nginx/njs.git && git -C njs checkout ${NJS_VER}"; \
    } > /usr/share/doc/mesh-in-agent/UPSTREAM_SOURCES.txt

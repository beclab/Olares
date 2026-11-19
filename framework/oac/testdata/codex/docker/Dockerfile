# Codex CLI base image for Olares.
#
# Scope: Ubuntu 24.04 with common CLI deps, language runtimes / build
# toolchains (Python / Node / Go / Rust / Java / Ruby / PHP / Lua),
# database clients, protobuf, Gradle, and extra -dev headers for
# typical agent-driven dev work. Non-root user (UID 1000) runs the Pod.
#
# The codex binary itself is NOT baked into this image. It is installed
# at runtime by the init-install-codex initContainer into the shared
# /opt/data PVC (under ~/.local/bin/codex + ~/.codex/).
#
# ripgrep is NOT installed here: codex ships its own bundled search and
# installing the system one alongside can confuse its file search.
#
# Build & push:
#   docker buildx build \
#     --platform linux/amd64,linux/arm64 \
#     -t beclab/harveyff-codex-base:0.5.9 \
#     --push .
# syntax=docker/dockerfile:1.7

# ---- Stage 1: olares-cli binary + Olares skills (Codex layout) ----
# @olares/cli postinstall downloads the Go binary; copy vendor binary (not JS
# wrapper) so the full verb set is available. skills CLI needs git to clone.
FROM node:24-bookworm-slim AS olares-assets
ARG OLARES_CLI_NPM_VERSION=1.12.7-cli.0

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      git ca-certificates \
 && rm -rf /var/lib/apt/lists/*

RUN npm install -g @olares/cli@${OLARES_CLI_NPM_VERSION} \
 && install -D "$(npm root -g)/@olares/cli/vendor/olares-cli" /out/bin/olares-cli \
 && ( /out/bin/olares-cli --version || /out/bin/olares-cli version || true )

WORKDIR /skills-build
# Codex project-local layout is .agents/skills (global is $CODEX_HOME/skills).
RUN npx -y skills add beclab/Olares -a codex -y --copy \
 && test -d /skills-build/.agents/skills \
 && mkdir -p /out/skills \
 && cp -a /skills-build/.agents/skills/. /out/skills/ \
 && ls -la /out/skills

# ---- Stage 2: Ubuntu runtime (no Node in final image) ----
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl \
 && mkdir -p -m 755 /etc/apt/keyrings \
 && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      -o /etc/apt/keyrings/githubcli-archive-keyring.gpg \
 && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
 && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        git-lfs \
        gh \
        openssh-client \
        jq \
        yq \
        less \
        locales \
        procps \
        tzdata \
        unzip \
        zip \
        vim-tiny \
        wget \
        xz-utils \
        pigz \
        file \
        rsync \
        patch \
        pkg-config \
        build-essential \
        cmake \
        ninja-build \
        clang \
        clang-format \
        clang-tidy \
        python3 \
        python3-dev \
        python3-pip \
        python3-venv \
        pipx \
        golang-go \
        rustc \
        cargo \
        ruby \
        ruby-dev \
        ruby-bundler \
        openjdk-21-jdk-headless \
        maven \
        gradle \
        php8.3-cli \
        php8.3-curl \
        php8.3-intl \
        php8.3-mbstring \
        php8.3-mysql \
        php8.3-pgsql \
        php8.3-sqlite3 \
        php8.3-xml \
        php8.3-zip \
        composer \
        lua5.4 \
        sqlite3 \
        gfortran \
        postgresql-client \
        default-mysql-client \
        redis-tools \
        protobuf-compiler \
        libprotobuf-dev \
        libssl-dev \
        libffi-dev \
        zlib1g-dev \
        libbz2-dev \
        libreadline-dev \
        libsqlite3-dev \
        liblzma-dev \
        libxml2-dev \
        libxslt1-dev \
        libyaml-dev \
        libjpeg-turbo8-dev \
        libpng-dev \
        libcurl4-openssl-dev \
        libicu-dev \
        libldap2-dev \
        libpq-dev \
        default-libmysqlclient-dev \
        gettext \
        perl \
        tmux \
        htop \
        dnsutils \
        iproute2 \
        iputils-ping \
        lsof \
        net-tools \
        traceroute \
        whois \
        inetutils-telnet \
        tcpdump \
        netcat-openbsd \
        socat \
        xmlstarlet \
        shellcheck \
        graphviz \
        moreutils \
        tree \
        nano \
        lrzsz \
        zstd \
        bzip2 \
        bash-completion \
        man-db \
        strace \
        ncdu \
        parallel \
        time \
        gdb \
        ltrace \
        valgrind \
        mtr-tiny \
        expect \
        aria2 \
        httpie \
 && ln -sfn "/usr/lib/jvm/java-21-openjdk-$(dpkg --print-architecture)" /usr/lib/jvm/java-21-openjdk-current \
 && rm -rf /var/lib/apt/lists/* \
 && locale-gen C.UTF-8 \
 && (userdel -r ubuntu 2>/dev/null || true) \
 && (groupdel ubuntu 2>/dev/null || true) \
 && useradd --create-home --uid 1000 --shell /bin/bash codex \
 && mkdir -p /opt/data \
 && chown -R 1000:1000 /opt/data

# Node 20 LTS via NodeSource. Ubuntu 24.04's nodejs is v18.19.1, which is
# too old for some CLIs (e.g. @larksuite/cli needs node >= 20.12.0). This
# ships node + npm under /usr/local, replacing the distro packages above.
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
 && apt-get install -y --no-install-recommends nodejs \
 && rm -rf /var/lib/apt/lists/* \
 && node --version && npm --version

COPY --from=olares-assets /out/bin/olares-cli /usr/local/bin/olares-cli
RUN chmod 755 /usr/local/bin/olares-cli \
 && ( /usr/local/bin/olares-cli --version || /usr/local/bin/olares-cli version || true )

# Baked skills; init-skills copies into $HOME/.codex/skills on the PVC.
COPY --from=olares-assets /out/skills /skills-staging
RUN chown -R 1000:1000 /skills-staging

ENV JAVA_HOME=/usr/lib/jvm/java-21-openjdk-current

USER 1000:1000
WORKDIR /opt/data

ENV HOME=/opt/data \
    PATH=/opt/data/.local/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    TERM=xterm-256color \
    PIP_BREAK_SYSTEM_PACKAGES=1 \
    NPM_CONFIG_PREFIX=/opt/data/.local \
    NPM_CONFIG_CACHE=/opt/data/.npm

CMD ["bash"]

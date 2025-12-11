# This file is auto generated from it's template,
# see citusdata/tools/packaging_automation/templates/docker/latest/latest.tmpl.dockerfile.
FROM postgres:17.5
ARG VERSION=13.0.3
LABEL maintainer="Citus Data https://citusdata.com" \
      org.label-schema.name="Citus" \
      org.label-schema.description="Scalable PostgreSQL for multi-tenant and real-time workloads" \
      org.label-schema.url="https://www.citusdata.com" \
      org.label-schema.vcs-url="https://github.com/citusdata/citus" \
      org.label-schema.vendor="Citus Data, Inc." \
      org.label-schema.version=${VERSION} \
      org.label-schema.schema-version="1.0"

ENV CITUS_VERSION ${VERSION}.citus-1

# install Citus
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       ca-certificates \
       curl \
       lsb-release \
       postgresql-17-pgvector 

COPY docker/citus/deb.sh /tmp/deb.sh
RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    curl -sfLO https://repo.pigsty.io/apt/pgsql/bookworm/pool/main/c/citus/postgresql-17-citus_13.0.3-9PIGSTY~bookworm_${arch}.deb && \
    apt-get install ./postgresql-17-citus_13.0.3-9PIGSTY~bookworm_${arch}.deb && rm postgresql-17-citus_13.0.3-9PIGSTY~bookworm_${arch}.deb

RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    curl -sfLO https://apt.postgresql.org/pub/repos/apt/pool/main/p/postgresql-hll/postgresql-17-hll_2.18-2.pgdg120+1_${arch}.deb && \
    apt-get install ./postgresql-17-hll_2.18-2.pgdg120+1_${arch}.deb && rm postgresql-17-hll_2.18-2.pgdg120+1_${arch}.deb

RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    curl -sfLO https://repo.pigsty.io/apt/pgsql/bookworm/pool/main/t/topn/postgresql-17-topn_2.7.0-1PIGSTY~bookworm_${arch}.deb && \
    apt-get install ./postgresql-17-topn_2.7.0-1PIGSTY~bookworm_${arch}.deb && rm postgresql-17-topn_2.7.0-1PIGSTY~bookworm_${arch}.deb

    # bash /tmp/deb.sh $arch \
    # && apt-get install -y postgresql-$PG_MAJOR-citus \
    #                       postgresql-$PG_MAJOR-hll \
    #                       postgresql-$PG_MAJOR-topn 

RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    curl -sfLO https://github.com/tensorchord/pgvecto.rs/releases/download/v0.4.0/vectors-pg17_0.4.0_${arch}.deb && \
    apt-get install ./vectors-pg17_0.4.0_${arch}.deb && rm vectors-pg17_0.4.0_${arch}.deb

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       build-essential \
       postgresql-server-dev-17 \
       lbzip2 \
        && curl -sfLO http://www.xunsearch.com/scws/down/scws-1.2.3.tar.bz2  \
        && tar xf scws-1.2.3.tar.bz2 \
        && cd scws-1.2.3 \
        && ./configure \
        && make install \
        && echo "/usr/local/lib" | tee /etc/ld.so.conf.d/scws.conf \
        && ldconfig \
        && apt-get install -y equivs \
        && equivs-control scws \
        && echo "----------------- before -----------------" \
        && cat scws \
        && sed -i 's/^# Version:.*/Version: 1.2.3/' scws \
        && sed -i 's/^Package:.*/Package: scws/' scws \
        && echo "----------------- after -----------------" \
        && cat scws \
        && equivs-build scws \
        && ls -la *.deb \
        && dpkg -i scws_1.2.3_all.deb \
        && apt-get install -y locales-all \
        && locale-gen zh_CN.UTF-8 \
        && apt-get purge -y --auto-remove build-essential postgresql-server-dev-17 libpq-dev \
        && rm -rf scws-1.2.3.tar.bz2 scws-1.2.3 


RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    curl -sfLO https://repo.pigsty.io/apt/pgsql/bookworm/pool/main/z/zhparser/postgresql-17-zhparser_2.3-1PIGSTY~bookworm_${arch}.deb && \
    apt-get install ./postgresql-17-zhparser_2.3-1PIGSTY~bookworm_${arch}.deb && rm postgresql-17-zhparser_2.3-1PIGSTY~bookworm_${arch}.deb
    

# RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
#     curl -sfLO https://apt.postgresql.org/pub/repos/apt/pool/main/p/postgis/postgresql-17-postgis-3_3.5.2+dfsg-1.pgdg120+1_${arch}.deb && \
#     curl -sfLO https://apt.postgresql.org/pub/repos/apt/pool/main/p/postgis/postgresql-17-postgis-3-scripts_3.5.2+dfsg-1.pgdg120+1_${arch}.deb && \

RUN apt-get update \
&& apt-get install -y --no-install-recommends \
postgresql-17-postgis-3 postgresql-17-postgis-3-scripts 

RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) && \
    curl -sfLO https://github.com/tensorchord/VectorChord/releases/download/0.3.0/postgresql-17-vchord_0.3.0-1_$(dpkg --print-architecture).deb && \
    apt install ./postgresql-17-vchord_0.3.0-1_$(dpkg --print-architecture).deb && rm postgresql-17-vchord_0.3.0-1_$(dpkg --print-architecture).deb

# add citus to default PostgreSQL config
RUN echo "shared_preload_libraries='citus,vectors.so,vchord.so'" >> /usr/share/postgresql/postgresql.conf.sample

# add scripts to run after initdb
COPY docker/citus/001-create-citus-extension.sql /docker-entrypoint-initdb.d/

# add health check script
COPY docker/citus/pg_healthcheck docker/citus/wait-for-manager.sh /
RUN chmod +x /wait-for-manager.sh

# entry point unsets PGPASSWORD, but we need it to connect to workers
# https://github.com/docker-library/postgres/blob/33bccfcaddd0679f55ee1028c012d26cd196537d/12/docker-entrypoint.sh#L303
RUN sed "/unset PGPASSWORD/d" -i /usr/local/bin/docker-entrypoint.sh

HEALTHCHECK --interval=4s --start-period=6s CMD ./pg_healthcheck

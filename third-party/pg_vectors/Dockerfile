FROM citusdata/citus:11.3.0

RUN apt-get update
RUN apt-get install -y wget

RUN wget https://github.com/tensorchord/pgvecto.rs/releases/download/v0.2.1/vectors-pg15_0.2.1_amd64.deb

RUN apt-get install ./vectors-pg15_0.2.1_amd64.deb
RUN rm -f ./vectors-pg15_0.2.1_amd64.deb

RUN echo "shared_preload_libraries='citus, vectors.so'" >> /usr/share/postgresql/postgresql.conf.sample

# escape=\
RUN echo "search_path='\"\$user\", public, vectors'" >> /usr/share/postgresql/postgresql.conf.sample


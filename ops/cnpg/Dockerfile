# ---------- Base builder image ----------
FROM postgres:17.6-bookworm AS base-build
SHELL ["/bin/bash", "-lc"]

ARG PG_MAJOR=17
ARG ICU_PKG=libicu72

# Core build deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates git curl wget gnupg pkg-config jq \
    libssl-dev libclang-dev llvm-dev clang \
    postgresql-server-dev-${PG_MAJOR} libpq-dev \
    ${ICU_PKG} \
    && rm -rf /var/lib/apt/lists/*

# Rust (shared across stages)
ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y \
    && . /usr/local/cargo/env \
    && rustup default stable

# ---------- Build pgvector (C / PGXS) ----------
FROM base-build AS build-pgvector
ARG PG_MAJOR=17
ARG PGVECTOR_VERSION=v0.8.1

WORKDIR /tmp
RUN git clone --branch ${PGVECTOR_VERSION} --depth 1 https://github.com/pgvector/pgvector.git
WORKDIR /tmp/pgvector
RUN make && make install

# ---------- Build pgvectorscale (Rust / pgrx 0.12.x) ----------
FROM base-build AS build-pgvectorscale
ARG PG_MAJOR=17
ARG PGVECTOR_SCALE_VERSION=0.8.0

# Install cargo-pgrx 0.12.x
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    cargo install --locked cargo-pgrx --version 0.12.9
RUN cargo pgrx init --pg${PG_MAJOR} /usr/lib/postgresql/${PG_MAJOR}/bin/pg_config

WORKDIR /tmp
RUN git clone --branch ${PGVECTOR_SCALE_VERSION} --depth 1 https://github.com/timescale/pgvectorscale.git
WORKDIR /tmp/pgvectorscale/pgvectorscale
RUN cargo pgrx install --release


# ---------- Build ParadeDB (Rust / pgrx 0.15.x) ----------
FROM base-build AS build-paradedb
ARG PG_MAJOR=17
ARG PARADEDB_VERSION=v0.17.12

# Install cargo-pgrx 0.15.x
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    cargo install --locked cargo-pgrx --version 0.15.0

RUN cargo pgrx init --pg${PG_MAJOR} /usr/lib/postgresql/${PG_MAJOR}/bin/pg_config

WORKDIR /tmp
RUN git clone --branch ${PARADEDB_VERSION} --depth 1 https://github.com/paradedb/paradedb.git
WORKDIR /tmp/paradedb/pg_search
RUN cargo pgrx install --release --features icu


# ---------- Runtime ----------
FROM postgres:17.6-bookworm
ARG PG_MAJOR=17
ARG ICU_PKG=libicu72

# Runtime deps only
RUN apt-get update && apt-get install -y --no-install-recommends \
    ${ICU_PKG} \
    && rm -rf /var/lib/apt/lists/*

# Copy pgvector
COPY --from=build-pgvector /usr/lib/postgresql/${PG_MAJOR}/lib/vector*.so /usr/lib/postgresql/${PG_MAJOR}/lib/
COPY --from=build-pgvector /usr/share/postgresql/${PG_MAJOR}/extension/vector* /usr/share/postgresql/${PG_MAJOR}/extension/

# Copy pgvectorscale
COPY --from=build-pgvectorscale /usr/lib/postgresql/${PG_MAJOR}/lib/vectorscale*.so /usr/lib/postgresql/${PG_MAJOR}/lib/
COPY --from=build-pgvectorscale /usr/share/postgresql/${PG_MAJOR}/extension/vectorscale* /usr/share/postgresql/${PG_MAJOR}/extension/

# Copy ParadeDB (pg_search etc.)
COPY --from=build-paradedb /usr/lib/postgresql/${PG_MAJOR}/lib/pg_search*.so /usr/lib/postgresql/${PG_MAJOR}/lib/
COPY --from=build-paradedb /usr/share/postgresql/${PG_MAJOR}/extension/pg_search* /usr/share/postgresql/${PG_MAJOR}/extension/

# Optional: strip binaries to reduce size
RUN find /usr/lib/postgresql/${PG_MAJOR}/lib -name '*.so' -exec strip --strip-unneeded {} + || true

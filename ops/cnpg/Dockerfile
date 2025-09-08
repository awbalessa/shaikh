# ---------- Builder ----------
FROM postgres:17.6-bookworm AS build
SHELL ["/bin/bash", "-lc"]

# Core build deps (C, headers, Rust tooling, ICU for tokenizers)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates git curl wget gnupg pkg-config jq \
    libssl-dev libclang-dev llvm-dev clang \
    postgresql-server-dev-17 libpq-dev \
    libicu70 \
    && rm -rf /var/lib/apt/lists/*

# ---- Install Rust (needed for pgrx extensions)
ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y \
    && rustup default stable

# ---- Versions (overridable at build time)
ARG PG_MAJOR=17
ARG PGVECTOR_VERSION=v0.8.1
ARG PGVECTOR_SCALE_VERSION=v0.8.0
ARG PARADEDB_VERSION=v0.17.12

# ---- pgvector (C / PGXS)
WORKDIR /tmp
RUN git clone --branch ${PGVECTOR_VERSION} --depth 1 https://github.com/pgvector/pgvector.git
WORKDIR /tmp/pgvector
RUN make && make install

# ---- pgvectorscale (Rust / pgrx)
WORKDIR /tmp
RUN git clone --branch ${PGVECTOR_SCALE_VERSION} --depth 1 https://github.com/timescale/pgvectorscale.git
WORKDIR /tmp/pgvectorscale/pgvectorscale
RUN cargo install --locked cargo-pgrx --version \
    "$(cargo metadata --format-version 1 | jq -r '.packages[] | select(.name==\"pgrx\") | .version')"
RUN cargo pgrx init --pg${PG_MAJOR} /usr/lib/postgresql/${PG_MAJOR}/bin/pg_config
RUN cargo pgrx install --release

# ---- ParadeDB (Rust / pgrx, ICU-enabled)
WORKDIR /tmp
RUN git clone --branch ${PARADEDB_VERSION} --depth 1 https://github.com/paradedb/paradedb.git
WORKDIR /tmp/paradedb
RUN cargo install --locked cargo-pgrx --version \
    "$(cargo metadata --format-version 1 | jq -r '.packages[] | select(.name==\"pgrx\") | .version')"
RUN cargo pgrx init --pg${PG_MAJOR} /usr/lib/postgresql/${PG_MAJOR}/bin/pg_config
RUN cargo pgrx install --release --features icu

# ---------- Runtime ----------
FROM postgres:17.6-bookworm

# Install runtime deps (ICU required for pg_search tokenizer)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libicu70 \
    && rm -rf /var/lib/apt/lists/*

# Copy compiled artifacts from builder
COPY --from=build /usr/lib/postgresql/17/lib/ /usr/lib/postgresql/17/lib/
COPY --from=build /usr/share/postgresql/17/extension/ /usr/share/postgresql/17/extension/
COPY --from=build /usr/include/postgresql/server/ /usr/include/postgresql/server/

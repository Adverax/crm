FROM postgres:16-alpine

# Build-time dependencies for pgTAP compilation
RUN apk add --no-cache --virtual .build-deps \
        make \
        gcc \
        musl-dev \
        perl-dev \
        git \
        patch \
    && apk add --no-cache \
        perl \
        perl-app-cpanminus \
    # Build pgTAP from source
    && git clone --branch v1.3.3 --depth 1 https://github.com/theory/pgtap.git /tmp/pgtap \
    && cd /tmp/pgtap \
    && make \
    && make install \
    # Install pg_prove (TAP::Parser::SourceHandler::pgTAP)
    && cpanm --notest TAP::Parser::SourceHandler::pgTAP \
    # Cleanup
    && rm -rf /tmp/pgtap \
    && apk del .build-deps

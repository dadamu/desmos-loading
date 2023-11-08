FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates

# Install bash
RUN apk add --no-cache bash

# Copy over binaries from the build-env
COPY --from=kilem/builder:latest /code/build/desmos-loading /usr/bin/desmos-loading


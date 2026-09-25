# Build the static binary. Dependencies resolve in their own layer, so a
# source-only change does not re-download the module graph.
FROM golang:1.26 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build metadata for the linker. The Makefile derives these from git; here the
# caller supplies them:
#   docker build --build-arg VERSION=$(git describe --tags --always --dirty) .
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w \
        -X github.com/example/mycli/internal/cli.Version=${VERSION} \
        -X github.com/example/mycli/internal/cli.Commit=${COMMIT} \
        -X github.com/example/mycli/internal/cli.BuildDate=${BUILD_DATE}" \
      -o /out/mycli ./cmd/mycli

# Runtime. scratch carries nothing but the binary: no shell, no package
# manager, and nothing to patch.
FROM scratch

# The binary makes no network calls today. The trust store is here so the first
# command that does will not fail TLS verification for a missing CA bundle.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /out/mycli /mycli

# Run unprivileged. scratch has no /etc/passwd, so the user is numeric.
USER 65532:65532

ENTRYPOINT ["/mycli"]

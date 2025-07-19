# ============ Stage 1: Builder ============
FROM golang:1.24-alpine AS builder

# args
ARG PKG="github.com/gitslim/go-ragger"
ARG NAME="server"
ARG BUILD_VERSION="v1.0.0"

# install git
RUN apk add --no-cache git

# install tools
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
      go install github.com/go-task/task/v3/cmd/task@v3.44.0 && \
      go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 && \
      go install github.com/a-h/templ/cmd/templ@v0.3.898
      
WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# run code generation
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    task generate

# set build info
RUN BUILD_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") && \
    BUILD_DATE=$(date +%Y%m%d) && \
    export LDFLAGS="-X ${PKG}/internal/version.buildVersion=${BUILD_VERSION} \
                    -X ${PKG}/internal/version.buildDate=${BUILD_DATE} \
                    -X ${PKG}/internal/version.buildCommit=${BUILD_COMMIT}"

# build app
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="${LDFLAGS} -w -s" \
    -o /app/${NAME} \
    ./cmd/${NAME}

# ============ Stage 2: Runtime ============
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/server .
COPY --chown=nonroot:nonroot ./static ./static

USER nonroot:nonroot
EXPOSE 8888
ENTRYPOINT ["/app/server"]

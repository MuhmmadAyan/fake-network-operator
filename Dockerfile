# Stage 1: Build binaries
FROM golang:1.23-alpine AS builder

WORKDIR /workspace

# Copy dependency manifests and download modules
COPY go.mod go.sum* ./
RUN go mod download

# Copy source directory
COPY . .

# Create output directory
RUN mkdir -p /out

# Build all binaries from cmd/ with CGO disabled for linux target
RUN for dir in $(ls ./cmd); do \
      outname="$dir"; \
      case "$dir" in \
        nic-labeler) outname="fake-nic-labeler" ;; \
        rdma-device-plugin) outname="fake-rdma-device-plugin" ;; \
        sriov-device-plugin) outname="fake-sriov-device-plugin" ;; \
        sriov-config-daemon) outname="fake-sriov-config-daemon" ;; \
        nic-status-exporter) outname="fake-nic-status-exporter" ;; \
        cli-injector) outname="fake-cli-injector" ;; \
        ib-kubernetes) outname="fake-ib-kubernetes" ;; \
        ufm-stub) outname="fake-ufm-stub" ;; \
        nic-dra-plugin) outname="fake-nic-dra-plugin" ;; \
      esac; \
      echo "Building ./cmd/$dir as /out/$outname..." && \
      CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/$outname ./cmd/$dir ; \
    done

# Stage 2: Minimal static runtime image
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /out/ /usr/local/bin/

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/controller-manager"]

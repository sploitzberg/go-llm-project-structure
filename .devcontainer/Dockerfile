FROM golang:1.26.5@sha256:ae5a2316d12f3e78fd99177dad452e6ad4f240af2d71d57b480c3477f250fec6

# Install basic tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Create the non-root development user
RUN useradd -m -s /bin/bash vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install Task (task runner)
RUN go install github.com/go-task/task/v3/cmd/task@v3.52.0

# Install RTK for token-efficient Zed tasks
ARG TARGETARCH=amd64
ARG RTK_VERSION=0.43.0
RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) \
            rtk_asset="rtk-x86_64-unknown-linux-musl.tar.gz"; \
            rtk_checksum="ff8a1e7766496e175291a85aeca1dc97c9ff6df33e51e5893d1fbc78fea2a609" \
            ;; \
        arm64) \
            rtk_asset="rtk-aarch64-unknown-linux-gnu.tar.gz"; \
            rtk_checksum="5519f7ca12e5c143a609f0d28a0a77b97413a8dce31c2681f1a41c24519a8731" \
            ;; \
        *) \
            echo "Unsupported architecture: ${TARGETARCH}" >&2; \
            exit 1 \
            ;; \
    esac; \
    curl -fsSLo /tmp/rtk.tar.gz "https://github.com/rtk-ai/rtk/releases/download/v${RTK_VERSION}/${rtk_asset}"; \
    echo "${rtk_checksum}  /tmp/rtk.tar.gz" | sha256sum -c -; \
    tar -xzf /tmp/rtk.tar.gz -C /usr/local/bin rtk; \
    rm /tmp/rtk.tar.gz

# Install the same pinned tools used by local and hosted CI
COPY scripts/setup/install-ci-tools.sh /tmp/install-ci-tools.sh
RUN bash /tmp/install-ci-tools.sh && rm /tmp/install-ci-tools.sh

# Install Go tools for Zed
RUN go install golang.org/x/tools/gopls@v0.23.0 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.27.0

# Set GOPATH to user home for proper permissions
ENV GOPATH="/home/vscode/go"
ENV PATH="/home/vscode/go/bin:${PATH}"

# Create Go directory structure in user home
RUN mkdir -p /home/vscode/go/{bin,pkg,src} && \
    chown -R vscode:vscode /home/vscode/go

# Set working directory
WORKDIR /workspace

# Switch to vscode user
USER vscode

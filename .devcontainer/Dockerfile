FROM golang:1.26

# Install basic tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    jq \
    sudo \
    && rm -rf /var/lib/apt/lists/*

# Create vscode user
RUN useradd -m -s /bin/bash vscode && \
    echo "vscode ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install Task (task runner)
RUN go install github.com/go-task/task/v3/cmd/task@latest

# Install golangci-lint
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Install yq (YAML processor)
RUN wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && \
    chmod +x /usr/local/bin/yq

# Install Go tools for CI
RUN go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0 && \
    go install github.com/loov/goda@latest

# Set working directory
WORKDIR /workspace

# Switch to vscode user
USER vscode

# Set PATH
ENV PATH="/go/bin:${PATH}"

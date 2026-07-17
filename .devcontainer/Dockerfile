FROM golang:1.26

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
RUN go install github.com/go-task/task/v3/cmd/task@latest

# Install golangci-lint
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Install Go tools for Zed and CI
RUN go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0 && \
    go install github.com/loov/goda@latest

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

# Docker

The bootstrap image runs the removable greeting CLI example. It does not expose a network port.

## Build and run

```bash
docker build -t go-llm-project-structure .
docker run --rm go-llm-project-structure Go Developer
```

Expected output:

```text
Hello, Go Developer!
```

The Compose example supplies `Docker` as the name:

```bash
docker compose up --build
```

To pass different arguments through Compose:

```bash
docker compose run --rm app Go Developer
```

When replacing the greeting CLI with an HTTP or gRPC adapter, update the Docker `ENTRYPOINT`, Compose service, and exposed ports together.

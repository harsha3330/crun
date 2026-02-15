# crun

A minimal OCI-style container runtime in Go. Run container images with overlay filesystems, optional host networking, and clean lifecycle (run, stop, remove image).

## Requirements

- Linux (overlay, namespaces)
- Go 1.24+ (to build)
- Root for run/stop (mounts, network, chroot)

## Quick start

```bash
# One-time setup
./bin/crun init --log-level debug
./bin/crun pull nginx:1-alpine-perl

# Run (detached); get container-id from output
sudo ./bin/crun run --network-host nginx:1-alpine-perl

# View logs, stop, remove image
cat ~/.crun/containers/<container-id>/log
sudo ./bin/crun stop <container-id>
./bin/crun rmi nginx:1-alpine-perl
```

## Commands

| Command | Description |
|--------|-------------|
| `init` | Initialize crun (config, log settings). Run once. |
| `pull <image>` | Pull an image from Docker Hub (e.g. `nginx:1-alpine-perl`). |
| `run [--network-host] <image>` | Start a container (detached). Use `--network-host` to access UI at http://localhost. |
| `stop <container-id>` | Stop the container, unmount overlay, remove container dir. |
| `rmi <image>` | Remove a pulled image (tag + manifest). Blobs remain until prune. |
| `images` | List pulled images (repo:tag). |
| `ps` | List running containers (id, image, pid, status). |

See [docs/usage.md](docs/usage.md) for detailed usage and examples.

## Build

```bash
go build -o bin/crun ./cmd/crun
```

## Data layout

- **Config:** `~/.crun/config.toml` (after `init`)
- **Images:** `~/.crun/images/`, `~/.crun/blobs/`, `~/.crun/layers/`
- **Containers:** `~/.crun/containers/<id>/` (log, pid, overlay; removed on `stop`)

## License

See repository.

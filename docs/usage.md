# crun usage and examples

## Initial setup

Run once per machine (or user config):

```bash
./bin/crun init
```

Optional: set log level and format.

```bash
./bin/crun init --log-level debug --log-format text
```

Config is written to `~/.crun/config.toml`. Logs go to a file under `$TMPDIR/crun` (or similar).

---

## Pulling images

Pull from Docker Hub. Use explicit tags (no `latest`).

```bash
./bin/crun pull nginx:1-alpine-perl
./bin/crun pull busybox:1.36
```

Images are stored under `~/.crun/images/`, `~/.crun/blobs/`, and `~/.crun/layers/`.

---

## Running containers

Containers start **detached**: the CLI exits and the process keeps running.

```bash
sudo ./bin/crun run nginx:1-alpine-perl
```

Example output:

```
✔ container started (detached) container-id=abc123 pid=12345 logs=.../containers/abc123/log
→ stop: crun stop abc123
→ logs: cat /path/to/containers/abc123/log
→ service: listening in container (e.g. port 80); use --network=host to access UI at http://localhost
```

### Host network (access UI)

To reach the app from the host (e.g. open nginx in a browser):

```bash
sudo ./bin/crun run --network-host nginx:1-alpine-perl
```

Then open **http://localhost** (port 80 for nginx). Only one container can use a given host port with `--network-host`.

Without `--network-host`, each container has its own network namespace and can bind to port 80 inside the container, but there is no port mapping to the host yet.

---

## Viewing logs

Logs are written to a file per container. The run output prints the path:

```bash
cat ~/.crun/containers/<container-id>/log
```

Or use the path shown in the “logs:” line (e.g. `cat /home/user/.crun/containers/abc123/log`).

---

## Stopping containers

Stop by container ID (from the run output):

```bash
sudo ./bin/crun stop <container-id>
```

This:

1. Sends SIGTERM, then SIGKILL if needed
2. Unmounts the overlay at `containers/<id>/merged`
3. Removes the whole `containers/<id>/` directory

The image and layers are not removed.

---

## Removing images

Remove a pulled image so it can no longer be used with `run`:

```bash
./bin/crun rmi nginx:1-alpine-perl
```

This deletes the tag and manifest. Blobs and extracted layers stay on disk; a future “prune” command could reclaim that space.

---

## Listing images and containers

**Image list** – show all pulled images (repo:tag):

```bash
./bin/crun images
```

**Container list** – show running containers (id, image ref, pid, status):

```bash
./bin/crun ps
```

The image reference (e.g. `nginx:1-alpine-perl`) is stored when you run a container so `ps` can display it.

---

## Summary

| Goal | Command |
|------|--------|
| Setup | `./bin/crun init` |
| Pull image | `./bin/crun pull <image:tag>` |
| Run (detached) | `sudo ./bin/crun run [--network-host] <image>` |
| View logs | `cat ~/.crun/containers/<id>/log` |
| Stop container | `sudo ./bin/crun stop <id>` |
| Remove image | `./bin/crun rmi <image:tag>` |
| List images | `./bin/crun images` |
| List containers | `./bin/crun ps` |

All run/stop operations require root (sudo) for overlay mount, chroot, and network.

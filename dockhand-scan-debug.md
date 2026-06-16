# Dockhand scan debug report

Date: 2026-06-13  
Dockhand host: `centos-10.lab`  
Host OS: CentOS Stream 10  
Docker: `Docker version 29.5.2, build 79eb04c`  
Image under test: `ghcr.io/lima3w/padduck-frontend:latest`

## Summary

Dockhand/Grype is failing because Docker 29.5.2 on `centos-10.lab`, using the containerd-backed `overlayfs` image store, exports the local Padduck frontend image as an incomplete Docker/OCI archive.

The archive produced by `docker save ghcr.io/lima3w/padduck-frontend:latest` contains only:

```text
blobs/
blobs/sha256/
blobs/sha256/4d54804d43d29d719562e062309efdb4c1642520dd451a7555a3abfdd0e4d900
index.json
manifest.json
oci-layout
```

But `manifest.json` references one config blob and ten layer blobs that are not present in the tar. Both Trivy and Grype fail when they consume this archive. Manual Dockhand-style tagging to `latest-dockhand-pending` reproduces the same failure, so the tag name is not the cause.

Other images, including `busybox:latest`, `nginx:1.31.1-trixie`, and a locally built minimal image derived from `nginx:1.31.1-trixie`, export complete archives and scan successfully. This makes the failure specific to the published Padduck frontend image artifact as represented locally by Docker, not a general Docker outage and not a scanner-only problem.

## Root cause assessment

Best current root-cause call: the published Padduck frontend artifact is being pulled as a single Docker schema-v2 manifest:

```json
{
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "digest": "sha256:4d54804d43d29d719562e062309efdb4c1642520dd451a7555a3abfdd0e4d900",
  "size": 2403
}
```

On this Docker 29/containerd image-store host, that artifact is locally runnable but not safely exportable: Docker's save/export path writes the manifest descriptor but omits the config and layer blobs. Dockhand/Grype's local Docker path depends on that export path, so it fails.

The strongest package-level fix is to republish the frontend image in a form Docker 29 exports correctly on this host: an OCI image index / manifest list for `linux/amd64`, rather than the current single Docker schema-v2 manifest. The local reproduction showed that a derived `nginx:1.31.1-trixie` image built as an OCI index with an attestation exported correctly.

This is not caused by disk pressure:

```text
df -h: root 457G size, 31G used, 427G available, 7% used
df -ih: root 229M inodes, 789K used, 228M free, 1% used
docker system df: Images 12.64GB, Build Cache 0B
```

## What is failing

The failing path is Docker local image export for:

```text
ghcr.io/lima3w/padduck-frontend:latest
ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending
```

Scanner failures:

```text
Trivy --input /scan/padduck-frontend-new.tar:
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
```

```text
Grype docker-archive:/scan/padduck-frontend-new.tar:
docker-archive: unable to provide image from tarball:
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
```

```text
Grype local pending tag:
docker: unable to provide image from tarball:
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
```

## Why it is failing

`docker save` creates a tar whose `manifest.json` references blobs that are absent.

Expected: the tar contains the image config plus every layer referenced by `manifest.json`.  
Actual: the tar contains only six archive entries and is 8.5 KB.

Evidence from the manifest-reference check:

```text
archive_entries 6
manifest_items 1
referenced_blobs 11
missing_count 11
MISSING blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76
MISSING blobs/sha256/c95640bb37f6885da02c8c3b6d8e2daca996dfaa99ba6ad46f1b60155fdaf504
MISSING blobs/sha256/d4dcde3aeeedc4744f6c3a6c1f6d51eb75e01955bbef0480617fb91c0522f211
MISSING blobs/sha256/c5a7565de4cfb2ece2cfa07cf5c08da0ef0d447f9eca7e386bb504bd279e6407
MISSING blobs/sha256/57b3fbf43092480ec04b52f6d3d94b4b0e016aa9792bde6d363e8aa956030bf5
MISSING blobs/sha256/80fa08b690ad61a19ed8ce2f58d50efcd25036b4e0619dec599e9bbb7dc0101e
MISSING blobs/sha256/41c9f6f909405b416a21da29f1f8b453d243d27a61fe7d06c202f273c7b25669
MISSING blobs/sha256/6376488be516899e64d9cfac0e1a6a931e5e7487aaa820047a3e30a9a2e7c877
MISSING blobs/sha256/4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1
MISSING blobs/sha256/80a8d0b6ac6a9be7697b66742b3e5c640e4d28f308a3fe6ce821aae1da657620
MISSING blobs/sha256/0a9864d563be639cc1e44ddde5e702526169856dd4ffcfcf994153ecb93c2d9a
manifest_references_missing_blobs
```

## Timeline of findings

1. Verified host access:
   ```text
   centos-10.lab
   NAME="CentOS Stream"
   VERSION="10 (Coughlan)"
   Docker version 29.5.2, build 79eb04c
   ```

2. Verified current Padduck image digest after fresh pull:
   ```text
   ghcr.io/lima3w/padduck-frontend latest sha256:4d54804d43d29d719562e062309efdb4c1642520dd451a7555a3abfdd0e4d900
   ```

3. Verified the container runs:
   ```text
   container-started
   /usr/share/nginx/html contains built frontend assets
   ```

4. Re-tested `docker save` after the fresh pull:
   ```text
   /tmp/padduck-frontend-new.tar size: 8.5K
   tar contains only manifest descriptor and metadata
   missing_count 11
   ```

5. Scanner tests:
   ```text
   Trivy tar input: failed, missing config blob
   Grype docker-archive: failed, missing config blob
   Grype Docker socket against latest: succeeded
   Trivy Docker socket against latest: failed, missing config blob through temporary Docker export
   Trivy remote registry source: succeeded, found Debian 13.5 packages and 230 vulnerabilities
   ```

6. Manual Dockhand pending tag:
   ```text
   docker tag latest latest-dockhand-pending: exit 0
   docker save pending tag: 8.5K tar
   missing_count 11
   Grype docker-archive pending: failed, missing config blob
   Grype Docker socket pending: failed, missing config blob
   ```

7. Base image comparison:
   ```text
   docker pull nginx:1.31.1-trixie: up to date
   docker save nginx:1.31.1-trixie: 63M tar
   archive_entries 19
   referenced_blobs 8
   missing_count 0
   Grype docker-archive:/scan/nginx-trixie.tar: grype_exit=0
   ```

8. Minimal derived image comparison:
   ```text
   docker build -t localhost/padduck-export-repro:copy /tmp/padduck-export-repro: exit 0
   descriptor mediaType: application/vnd.oci.image.index.v1+json
   docker save localhost/padduck-export-repro:copy: 61M tar
   archive_entries 19
   referenced_blobs 9
   missing_count 0
   ```

## Command results

### Image state

Command:

```bash
docker image ls --digests | grep -E 'padduck-frontend|nginx|busybox' || true
```

Result:

```text
ghcr.io/lima3w/padduck-frontend latest sha256:4d54804d43d29d719562e062309efdb4c1642520dd451a7555a3abfdd0e4d900 4d54804d43d2 27 minutes ago 212MB
nginx 1.31.1-trixie sha256:608a100c71651bf5b773c89083b4a1ad7ef4b2bd05d7a7e552271e03123692ad 608a100c7165 2 days ago 238MB
busybox latest sha256:fd8d9aa63ba2f0982b5304e1ee8d3b90a210bc1ffb5314d980eb6962f1a9715d fd8d9aa63ba2 4 weeks ago 6.74MB
```

Command:

```bash
docker image inspect ghcr.io/lima3w/padduck-frontend:latest --format '{{json .Descriptor}}' | python3 -m json.tool
```

Result:

```json
{
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "digest": "sha256:4d54804d43d29d719562e062309efdb4c1642520dd451a7555a3abfdd0e4d900",
    "size": 2403
}
```

Command:

```bash
docker buildx imagetools inspect ghcr.io/lima3w/padduck-frontend:latest
```

Result:

```text
Name:      ghcr.io/lima3w/padduck-frontend:latest
MediaType: application/vnd.docker.distribution.manifest.v2+json
Digest:    sha256:4d54804d43d29d719562e062309efdb4c1642520dd451a7555a3abfdd0e4d900
```

### Docker save

Command:

```bash
docker save ghcr.io/lima3w/padduck-frontend:latest -o /tmp/padduck-frontend-new.tar
ls -lh /tmp/padduck-frontend-new.tar
tar tf /tmp/padduck-frontend-new.tar | sed -n '1,80p'
```

Result:

```text
-rw-------. 1 lima3 lima3 8.5K Jun 13 00:37 /tmp/padduck-frontend-new.tar
blobs/
blobs/sha256/
blobs/sha256/4d54804d43d29d719562e062309efdb4c1642520dd451a7555a3abfdd0e4d900
index.json
manifest.json
oci-layout
```

### Published manifest

Command:

```bash
docker manifest inspect ghcr.io/lima3w/padduck-frontend:latest | python3 -m json.tool | sed -n '1,160p'
```

Result:

```json
{
    "schemaVersion": 2,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "config": {
        "mediaType": "application/vnd.docker.container.image.v1+json",
        "size": 11114,
        "digest": "sha256:06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76"
    },
    "layers": [
        "sha256:c95640bb37f6885da02c8c3b6d8e2daca996dfaa99ba6ad46f1b60155fdaf504",
        "sha256:d4dcde3aeeedc4744f6c3a6c1f6d51eb75e01955bbef0480617fb91c0522f211",
        "sha256:c5a7565de4cfb2ece2cfa07cf5c08da0ef0d447f9eca7e386bb504bd279e6407",
        "sha256:57b3fbf43092480ec04b52f6d3d94b4b0e016aa9792bde6d363e8aa956030bf5",
        "sha256:80fa08b690ad61a19ed8ce2f58d50efcd25036b4e0619dec599e9bbb7dc0101e",
        "sha256:41c9f6f909405b416a21da29f1f8b453d243d27a61fe7d06c202f273c7b25669",
        "sha256:6376488be516899e64d9cfac0e1a6a931e5e7487aaa820047a3e30a9a2e7c877",
        "sha256:4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1",
        "sha256:80a8d0b6ac6a9be7697b66742b3e5c640e4d28f308a3fe6ce821aae1da657620",
        "sha256:0a9864d563be639cc1e44ddde5e702526169856dd4ffcfcf994153ecb93c2d9a"
    ]
}
```

### Scanner matrix

Command:

```bash
docker run --rm -v /tmp:/scan:ro aquasec/trivy:latest image --scanners vuln --input /scan/padduck-frontend-new.tar
```

Result:

```text
exit=1
unable to open /scan/padduck-frontend-new.tar as a Docker image:
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
```

Command:

```bash
docker run --rm -v /tmp:/scan:ro anchore/grype:latest docker-archive:/scan/padduck-frontend-new.tar
```

Result:

```text
exit=1
docker-archive: unable to provide image from tarball:
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
```

Command:

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro anchore/grype:latest ghcr.io/lima3w/padduck-frontend:latest
```

Result:

```text
exit=0
Grype produced vulnerability results for Debian packages.
```

Command:

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro aquasec/trivy:latest image --scanners vuln ghcr.io/lima3w/padduck-frontend:latest
```

Result:

```text
exit=1
failed to analyze layer ... unable to populate ...
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
```

Command:

```bash
docker run --rm aquasec/trivy:latest image --scanners vuln --image-src remote ghcr.io/lima3w/padduck-frontend:latest
```

Result:

```text
exit=0
Detected OS family="debian" version="13.5"
Report Summary: ghcr.io/lima3w/padduck-frontend:latest (debian 13.5), Vulnerabilities 230
Total: 230 (UNKNOWN: 1, LOW: 119, MEDIUM: 77, HIGH: 31, CRITICAL: 2)
```

Remote scanner success was used only to prove the registry image is fetchable and analyzable. It is not the recommended fix path.

### Manual Dockhand pending tag

Command:

```bash
docker tag ghcr.io/lima3w/padduck-frontend:latest ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending
docker save ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending -o /tmp/padduck-frontend-pending.tar
```

Result:

```text
/tmp/padduck-frontend-pending.tar size: 8.5K
archive_entries 6
referenced_blobs 11
missing_count 11
```

Command:

```bash
docker run --rm -v /tmp:/scan:ro anchore/grype:latest docker-archive:/scan/padduck-frontend-pending.tar
```

Result:

```text
exit=1
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
```

Command:

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro anchore/grype:latest ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending
```

Result:

```text
exit=1
docker: unable to provide image from tarball:
file blobs/sha256/06125e255040a23b78f85ea63285f9fe855859b7e83250eb493964e84c1e7b76 not found in tar
oci-registry: MANIFEST_UNKNOWN for latest-dockhand-pending
```

### Base image comparison

Command:

```bash
docker pull nginx:1.31.1-trixie
docker save nginx:1.31.1-trixie -o /tmp/nginx-trixie.tar
tar tf /tmp/nginx-trixie.tar | sed -n '1,80p'
```

Result:

```text
Digest: sha256:608a100c71651bf5b773c89083b4a1ad7ef4b2bd05d7a7e552271e03123692ad
/tmp/nginx-trixie.tar size: 63M
archive_entries 19
referenced_blobs 8
missing_count 0
```

Command:

```bash
docker run --rm -v /tmp:/scan:ro anchore/grype:latest docker-archive:/scan/nginx-trixie.tar
```

Result:

```text
grype_exit=0
Grype produced vulnerability results for Debian packages.
```

### Storage and Docker image store

Command:

```bash
docker info | sed -n '/Storage Driver/,+30p'
```

Result:

```text
Storage Driver: overlayfs
 driver-type: io.containerd.snapshotter.v1
containerd version: 193637f7ee8ae5f5aa5248f49e7baa3e6164966e
Kernel Version: 6.12.0-233.el10.x86_64
Operating System: CentOS Stream 10 (Coughlan)
Architecture: x86_64
```

Command:

```bash
docker system df
df -h
df -ih
```

Result:

```text
Images: 33 total, 24 active, 12.64GB, 625.3MB reclaimable
Containers: 27 total, 26 active, 105.1MB
Local Volumes: 11 total, 2GB
Build Cache: 0B
Root filesystem: 457G size, 31G used, 427G available, 7% used
Root inodes: 229M total, 789K used, 228M free, 1% used
```

### Image history

Command:

```bash
docker image history --no-trunc --format '{{.ID}}\t{{.CreatedSince}}\t{{.Size}}\t{{.CreatedBy}}\t{{.Comment}}' ghcr.io/lima3w/padduck-frontend:latest | sed -n '1,30p'
```

Important result:

```text
CMD ["nginx" "-g" "daemon off;"]                                0B
HEALTHCHECK curl -fs http://127.0.0.1:3000/health ...            0B
EXPOSE [3000/tcp]                                                0B
COPY nginx.conf /etc/nginx/conf.d/default.conf                   4.1kB
COPY /app/dist /usr/share/nginx/html                             4.58MB
RUN apt-get update && apt-get upgrade -y --no-install-recommends 0B
base nginx layers from nginx:1.31.1-trixie
```

The `apt-get upgrade` layer reports `0B`, so it does not appear to have changed package content in this build. The large vulnerability count mostly comes from the Debian/nginx runtime base, not from frontend application code.

## Safest fix

Fix the frontend image publication, not the scanner.

Recommended immediate package/build fix:

1. Publish the frontend image as an OCI image index / manifest list for `linux/amd64`.
2. Keep the image single-runtime-platform if production is amd64-only, but make the published artifact an OCI index so Docker 29/containerd exports a complete archive on the Dockhand host.
3. Verify the replacement image before rerunning Dockhand:

```bash
docker image rm ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending 2>/dev/null || true
docker image rm ghcr.io/lima3w/padduck-frontend:latest 2>/dev/null || true
docker pull ghcr.io/lima3w/padduck-frontend:latest
docker image inspect ghcr.io/lima3w/padduck-frontend:latest --format '{{json .Descriptor}}' | python3 -m json.tool
docker save ghcr.io/lima3w/padduck-frontend:latest -o /tmp/padduck-frontend-verify.tar
python3 - <<'PY'
import json, tarfile
path = "/tmp/padduck-frontend-verify.tar"
with tarfile.open(path) as tf:
    names = set(tf.getnames())
    manifest = json.load(tf.extractfile("manifest.json"))
missing = []
for item in manifest:
    cfg = item.get("Config")
    if cfg and cfg not in names:
        missing.append(cfg)
    for layer in item.get("Layers", []):
        if layer not in names:
            missing.append(layer)
print("missing_count", len(missing))
for item in missing:
    print("MISSING", item)
raise SystemExit(1 if missing else 0)
PY
docker run --rm -v /tmp:/scan:ro anchore/grype:latest docker-archive:/scan/padduck-frontend-verify.tar
```

Expected verification result:

```text
Descriptor mediaType should be application/vnd.oci.image.index.v1+json or another complete export-safe form.
docker save tar should be tens of MB, not 8.5K.
missing_count 0.
Grype docker-archive scan should exit 0.
```

## Recommended code/build/deploy changes

### Release workflow

Update both `.github/workflows/release.yml` and `.github/workflows/deploy.yml` for the frontend build so the published artifact is an OCI index for amd64.

Recommended frontend build-push settings:

```yaml
      - name: Build and push frontend
        uses: docker/build-push-action@10e90e3645eae34f1e60eeb005ba3a3d33f178e8  # v6.19.2
        with:
          context: ./frontend
          push: true
          platforms: linux/amd64
          provenance: mode=min
          sbom: true
          tags: ${{ steps.meta-frontend.outputs.tags }}
          labels: ${{ steps.meta-frontend.outputs.labels }}
```

Rationale:

- The local minimal derived image that exported correctly was an OCI image index and included an attestation manifest.
- The failing published image is a single Docker schema-v2 manifest.
- This changes the package artifact shape while keeping scanner behavior unchanged.

If SBOM publication creates unexpected registry compatibility issues, keep `platforms: linux/amd64` and `provenance: mode=min` first; the key target is an OCI index / complete export-safe artifact. Do not use scanner bypasses as the primary fix.

### Frontend Dockerfile

The current frontend runtime image is:

```dockerfile
FROM nginx:1.31.1-trixie
RUN apt-get update && apt-get upgrade -y --no-install-recommends && rm -rf /var/lib/apt/lists/*
```

Recommended cleanup:

```dockerfile
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD curl -fs http://127.0.0.1:3000/health | grep -q '"status":"ok"' || exit 1
CMD ["nginx", "-g", "daemon off;"]
```

If keeping Debian-based nginx, remove `apt-get upgrade` and instead rebuild regularly from a patched base image tag. Running distro upgrades inside the application Dockerfile makes the runtime less reproducible and does not appear to reduce the scanner count here; the history showed the upgrade layer as `0B`.

Important caveat: switching to `nginx:alpine` is good package hygiene and should reduce vulnerability noise, but by itself it is not proven to fix the Docker 29 export issue. The export issue is best addressed by publishing the image as an OCI index / manifest list and verifying `docker save` before Dockhand retry.

## Should Dockhand scan locally or remotely?

For the final fix, Dockhand can continue scanning locally after the image artifact is republished and verified to export correctly.

Remote scanning was useful as a diagnostic because it proved GHCR can serve the image and scanners can analyze the registry content. It should not be treated as the primary remediation for this incident because the goal is to fix the package/image artifact, and other images already work with Dockhand's local flow.

## Commands already run

Dockhand host commands run on `centos-10.lab`:

```bash
hostname
cat /etc/os-release | sed -n '1,6p'
docker --version
docker image ls --digests | grep -E 'padduck-frontend|nginx|busybox' || true
docker image inspect ghcr.io/lima3w/padduck-frontend:latest --format '{{json .Descriptor}}' | python3 -m json.tool
docker buildx imagetools inspect ghcr.io/lima3w/padduck-frontend:latest
docker run --rm ghcr.io/lima3w/padduck-frontend:latest sh -c 'echo container-started && ls -la /usr/share/nginx/html | head'
rm -f /tmp/padduck-frontend-new.tar
docker save ghcr.io/lima3w/padduck-frontend:latest -o /tmp/padduck-frontend-new.tar
ls -lh /tmp/padduck-frontend-new.tar
tar tf /tmp/padduck-frontend-new.tar | sed -n '1,80p'
tar -xOf /tmp/padduck-frontend-new.tar index.json | python3 -m json.tool
tar -xOf /tmp/padduck-frontend-new.tar manifest.json | python3 -m json.tool
python3 -c '...' # tar manifest reference checker
docker run --rm -v /tmp:/scan:ro aquasec/trivy:latest image --scanners vuln --input /scan/padduck-frontend-new.tar
docker run --rm -v /tmp:/scan:ro anchore/grype:latest docker-archive:/scan/padduck-frontend-new.tar
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro anchore/grype:latest ghcr.io/lima3w/padduck-frontend:latest
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro aquasec/trivy:latest image --scanners vuln ghcr.io/lima3w/padduck-frontend:latest
docker run --rm aquasec/trivy:latest image --scanners vuln --image-src remote ghcr.io/lima3w/padduck-frontend:latest
docker tag ghcr.io/lima3w/padduck-frontend:latest ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending
docker save ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending -o /tmp/padduck-frontend-pending.tar
ls -lh /tmp/padduck-frontend-pending.tar
tar tf /tmp/padduck-frontend-pending.tar | sed -n '1,80p'
python3 -c '...' # pending tar manifest reference checker
docker run --rm -v /tmp:/scan:ro anchore/grype:latest docker-archive:/scan/padduck-frontend-pending.tar
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro anchore/grype:latest ghcr.io/lima3w/padduck-frontend:latest-dockhand-pending
docker pull nginx:1.31.1-trixie
docker save nginx:1.31.1-trixie -o /tmp/nginx-trixie.tar
ls -lh /tmp/nginx-trixie.tar
tar tf /tmp/nginx-trixie.tar | sed -n '1,80p'
python3 -c '...' # nginx tar manifest reference checker
docker run --rm -v /tmp:/scan:ro anchore/grype:latest docker-archive:/scan/nginx-trixie.tar
docker info | sed -n '/Storage Driver/,+30p'
docker system df
df -h
df -ih
docker image inspect ghcr.io/lima3w/padduck-frontend:latest | python3 -m json.tool > /tmp/padduck-image-inspect.json
docker image history --no-trunc ghcr.io/lima3w/padduck-frontend:latest > /tmp/padduck-image-history.txt
docker image history --no-trunc --format '{{.ID}}\t{{.CreatedSince}}\t{{.Size}}\t{{.CreatedBy}}\t{{.Comment}}' ghcr.io/lima3w/padduck-frontend:latest | sed -n '1,30p'
docker image history --no-trunc --format '{{.ID}}\t{{.CreatedSince}}\t{{.Size}}\t{{.CreatedBy}}\t{{.Comment}}' nginx:1.31.1-trixie | sed -n '1,30p'
docker manifest inspect ghcr.io/lima3w/padduck-frontend:latest | python3 -m json.tool | sed -n '1,160p'
docker manifest inspect nginx:1.31.1-trixie | python3 -m json.tool | sed -n '1,160p'
rm -rf /tmp/padduck-export-repro && mkdir -p /tmp/padduck-export-repro
printf 'FROM nginx:1.31.1-trixie\nCOPY marker.txt /usr/share/nginx/html/marker.txt\n' > /tmp/padduck-export-repro/Dockerfile
printf 'marker\n' > /tmp/padduck-export-repro/marker.txt
docker build -t localhost/padduck-export-repro:copy /tmp/padduck-export-repro
docker image inspect localhost/padduck-export-repro:copy --format '{{json .Descriptor}}' | python3 -m json.tool
docker save localhost/padduck-export-repro:copy -o /tmp/padduck-export-repro-copy.tar
ls -lh /tmp/padduck-export-repro-copy.tar
tar tf /tmp/padduck-export-repro-copy.tar | sed -n '1,80p'
python3 -c '...' # local derived tar manifest reference checker
```

Dev repo commands run in `/home/lima3/padduck`:

```bash
pwd
rg --files -g 'AGENTS.md' -g '.claude/**' -g '.codex/**' -g '*dockhand*' -g '*Dockerfile*' -g '*README*'
git status --short --branch
sed -n '1,220p' README.md
sed -n '1,220p' frontend/Dockerfile
find . -maxdepth 3 ... docs/security/deploy discovery
sed -n '1,260p' .github/workflows/deploy.yml
sed -n '1,260p' .github/workflows/ci.yml
sed -n '1,220p' .github/wiki/Security.md
rg -n 'build-push-action|padduck-frontend|provenance|sbom|platforms|release' .github/workflows Makefile docker-compose.yml frontend -g '!frontend/node_modules/**'
sed -n '1,150p' .github/workflows/release.yml
```

## Unanswered questions

- Whether GHCR will preserve the desired OCI index shape after the next release build. This should be verified with:
  ```bash
  docker buildx imagetools inspect ghcr.io/lima3w/padduck-frontend:latest
  ```
- Whether Dockhand currently scans `latest` after pull or only the temporary pending tag. Manual pending-tag reproduction shows both fail today, so this does not block the package fix.
- Whether production requires non-amd64 frontend images. If yes, publish a true multi-platform index. If not, publish an amd64-only OCI index.

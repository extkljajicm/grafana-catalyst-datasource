# grafana-catalyst-datasource — Dev & Release Guide

> For user plugin docs, see [`src/README.md`](./src/README.md).

This guide covers local development, running a local Grafana, and creating releases with **`create_release.sh`** and **`create-git-release.sh`**. The project is scaffolded with **@grafana/create-plugin** (see `create-plugin-README.md` in the repo for scaffold details and commands).

---

## Prerequisites

- **Git**
- **Node.js ≥ 22** and **npm**
- **Go ≥ 1.21** (Go toolchain will auto-fetch newer if needed)
- **Docker** with the **Compose plugin**
- CLI tools for scripts: `zip`, `awk`, `sed`, `jq`

> Target: Grafana **12.1+**. Backend builds for multiple OS/ARCH via `Magefile.go`.

---

## Install & Dev

```bash
git clone https://github.com/extkljajicm/grafana-catalyst-datasource.git
cd grafana-catalyst-datasource
npm ci
```

Dev UI (hot reload):

```bash
npm run dev
```

Lint & tests:

```bash
npm run lint
npm run test
```

---

## Local Grafana (Docker)

A `docker-compose.yaml` is included to run a local Grafana with the plugin.

```bash
# rebuild containers when UI or backend changes
docker compose down
docker compose up -d --build

# force a clean re-create (e.g., after dependency updates)
docker compose up -d --force-recreate
```

Unsigned plugins are allowed in this dev setup. For production, follow Grafana’s signing guide.

---

## Git Workflow (quick)

```bash
git checkout -b feat/my-change
git add -A
git commit -m "feat: concise summary"
git push -u origin feat/my-change
```

Keep `CHANGELOG.md` up to date; release notes are derived from it.

---

## Build a Release (local)

1) **Package** (frontend + backend + zip):

```bash
./create_release.sh
```

Output: `grafana-catalyst-datasource-<version>.zip` in repo root.  
- If a Git **tag** (e.g., `v1.0.4`) exists, the script uses it as the version.  
- Otherwise it falls back to `package.json`.

2) **Tag the current commit**:

```bash
git tag -a v1.0.4 -m "Release v1.0.4"
```

3) **Publish to GitHub** (push branch + tag, and prepare notes from `CHANGELOG.md`):

```bash
./create-git-release.sh
```

The script pushes `main` and the tag, then generates `RELEASE_NOTES.md` from the matching section in `CHANGELOG.md`. Finalize the GitHub release in the UI and attach the zip if desired.

---

## CI Release from Git Tag

The GitHub Actions workflow updates the build-time version from the tag (e.g., `v1.0.4 → 1.0.4`) **before** building and packaging. To release via CI:

```bash
git add -A && git commit -m "release: prep v1.0.4"  # optional
git push origin main                                 # optional
git tag -a v1.0.4 -m "Release v1.0.4"
git push origin v1.0.4
```

CI will:
- Set `plugin.json.info.version` from the tag
- Build (frontend + backend via `Magefile.go`)
- Package and validate the artifact zip

---

## Troubleshooting

- **Plugin not visible**
  - Confirm Grafana **12.1+**
  - Check container logs for plugin load errors
  - Ensure `dist/` exists with `plugin.json` (placeholders like `%VERSION%`, `%TODAY%` are replaced during build)

- **Backend missing in package**
  - Ensure `Magefile.go` is present and CI uses `backend-target: buildAll`
  - Check that `plugin.json` has `"backend": true` and `"executable": "grafana-catalyst-datasource"`

- **Windows ADS file in zip**
  - Remove any `*:Zone.Identifier` file and add this to `.gitignore`:
    ```
    *:Zone.Identifier
    ```

---

## License

Apache-2.0 © extkljajicm — 2025-08-30

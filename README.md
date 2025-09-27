# grafana-catalyst-datasource — Dev & Release Guide

> For end‑user plugin docs, see [`src/README.md`](./src/README.md).  
> For scaffold details and common commands, see **`create-plugin-README.md`** in this repo.

This guide covers local development, running a local Grafana, and creating releases either **via GitHub Actions from a git tag** (CI path) or **locally** with `create_release.sh` (for test packaging).

---

## Prerequisites

- **Git**
- **Node.js ≥ 22** and **npm**
- **Go ≥ 1.21**
- **Docker** with the **Compose plugin**
- CLI tools: `zip`, `awk`, `sed`, `jq`

Target environment: Grafana **12.1+**. Backend binaries are built via `Magefile.go`.

---

## Install & Dev

```bash
sudo apt install -y git zip unzip gawk sed jq golang-go nodejs npm docker.io docker-compose
sudo usermod -aG docker $USER
```


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

A `docker-compose.yaml` is included to run a local Grafana with the plugin mounted.

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

Keep `CHANGELOG.md` updated. Commits should follow the [Conventional Commits](https://www.conventionalcommits.org/) specification to facilitate release note generation. The changelog is updated manually.

---

# Release

Choose **one** path below. Do **not** mix them.

## Option A — CI release from git tag (recommended)

Minimal commands (trigger CI by pushing a tag):

```bash
# 1) create and push the annotated tag
export V=1.0.5
git tag -a v$V -m "Release v$V"
git push origin v$V

# optional: push pending commits to main
git push origin main
```

What happens next (handled by `.github/workflows/release.yml`):
- The workflow reads the tag (`v1.0.5 → 1.0.5`) and injects it at build time.
- Runs `npm ci` and `npm run build` to produce `dist/` (including transformed `dist/plugin.json`).
- Builds the backend with `go run github.com/magefile/mage -v BuildAll` (currently **linux/amd64** only).
- Zips the plugin (`<plugin-id>-<version>.zip`) and **auto‑creates a GitHub Release** for the tag, attaching the ZIP (and `.sha1`).
  - To create a draft instead of publishing immediately, set `draft: true` in the workflow’s “Create GitHub Release” step.

> Note: Local scripts are **not used** in the CI path.

---

## Option B — Local/manual packaging (no CI)

Use this to build a ZIP locally for testing.

1) **(Recommended)** Tag first so the local packager uses the correct version:
```bash
export V=1.0.5
git tag -a v$V -m "Release v$V"
```

2) **Package locally** (frontend + backend + zip):
```bash
./create_release.sh
```

Output: `grafana-catalyst-datasource-<version>.zip` in the repo root.  


---

## Troubleshooting

- **Plugin not visible**
  - Confirm Grafana **12.1+**
  - Check container logs for plugin load errors
  - Ensure the build created `dist/` with processed `plugin.json` (placeholders like `%VERSION%` / `%TODAY%` are replaced during build)

- **Backend missing in package**
  - Ensure `Magefile.go` is present and used by the build
  - Confirm `plugin.json` has `"backend": true` and `"executable": "grafana-catalyst-datasource"`

- **Windows ADS file in zip**
  - Remove any `*:Zone.Identifier` file and add this to `.gitignore`:
    ```
    *:Zone.Identifier
    ```

---

## License

Apache-2.0 © extkljajicm

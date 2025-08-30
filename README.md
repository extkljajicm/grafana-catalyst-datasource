# grafana-catalyst-datasource — Dev & Release Guide

> For end‑user plugin docs, see [`src/README.md`](./src/README.md).  
> For scaffold details and common commands, see **`create-plugin-README.md`** in this repo.

This guide covers local development, running a local Grafana, and creating releases with **`create_release.sh`** and **`create-git-release.sh`**, or via GitHub Actions from a **git tag**.

---

## Prerequisites

- **Git**
- **Node.js ≥ 22** and **npm**
- **Go ≥ 1.21**
- **Docker** with the **Compose plugin**
- CLI tools for scripts: `zip`, `awk`, `sed`, `jq`

Target environment: Grafana **12.1+**. Backend binaries are built via `Magefile.go`.

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

Keep `CHANGELOG.md` updated; release notes are derived from it.

---

# Release

Choose **one** path below. Do **not** mix them.

## Option A — CI release from git tag (recommended)

Minimal commands (trigger CI by pushing a tag):

```bash
# 1) create and push the annotated tag (this is the only thing you MUST do)
export V=1.0.4
git tag -a v$V -m "Release v$V"
git push origin v$V

# optional: push pending commits to main
git push origin main
```

What happens next:
- GitHub Actions injects the tag version (`v1.0.4 → 1.0.4`) at build time (before webpack) so `dist/plugin.json` has the correct version.
- CI builds frontend + backend (via `Magefile.go`), validates, and packages the zip.
- Create the GitHub Release and attach the artifact from the workflow run (if not automated).

> Note: The **local scripts** (`create_release.sh`, `create-git-release.sh`) are **not used** in this CI path.

---

## Option B — Local/manual release (no CI packaging)

1) **Tag the current commit** (so local packaging uses the correct version):

```bash
export V=1.0.4
git tag -a v$V -m "Release v$V"
```

2) **Package locally** (frontend + backend + zip):

```bash
./create_release.sh
```

Output: `grafana-catalyst-datasource-<version>.zip` in the repo root.  
- If a Git **tag** (e.g., `v1.0.4`) exists, the script uses it as the version.  
- Otherwise it falls back to `package.json`.

3) **Publish to GitHub** (push branch + tag and prepare release notes from `CHANGELOG.md`):

```bash
./create-git-release.sh
```

This will:
- `git push origin main` and `git push origin v$V`
- Generate `RELEASE_NOTES.md` from your `CHANGELOG.md` for that tag

> If your changelog sections use the form `## [1.0.4] - YYYY-MM-DD` rather than `## v1.0.4`, adjust the script or add a matching `## v1.0.4` heading so notes are extracted.

---

## Troubleshooting

- **Plugin not visible**
  - Confirm Grafana **12.1+**
  - Check container logs for plugin load errors
  - Ensure the build created `dist/` with a processed `plugin.json` (placeholders like `%VERSION%` / `%TODAY%` are replaced during build)

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

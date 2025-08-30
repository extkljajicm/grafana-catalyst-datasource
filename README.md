# grafana-catalyst-datasource — Dev & Release Guide

> For user plugin docs, see [`src/README.md`](./src/README.md).

This document covers local development, running a Grafana test instance, and creating releases with the helper scripts
**`create_release.sh`** and **`create-git-release.sh`**.

---

## Prerequisites

- **Git**
- **Node.js ≥ 18** and **npm**
- **Go ≥ 1.24**
- **Docker** with the **Compose plugin**
- CLI tools used by the helper scripts: `zip`, `awk`, `sed`

> Tip: The plugin targets Grafana **12.1+** and builds a Linux/amd64 backend binary for runtime.

---

## Clone & install

```bash
git clone https://github.com/extkljajicm/grafana-catalyst-datasource.git
cd grafana-catalyst-datasource
npm ci
```

Optional quality checks:

```bash
npm run lint
npm run test
```

---

## Local Grafana (Docker)

A `docker-compose.yaml` is included for a quick, local test instance with the plugin mounted.

Common commands:

```bash
# rebuild containers when UI or backend changes
docker compose down
docker compose up -d --build

# force a clean re-create (e.g., after dependency updates)
docker compose up -d --force-recreate
```

Dev UI (hot reload):

```bash
npm run dev
```

> The development container permits **unsigned** plugins. For production, sign the plugin per Grafana’s guidelines.

---

## Working with Git

```bash
# create a feature branch
git checkout -b feat/my-change

# stage and commit
git add -A
git commit -m "feat: concise summary of the change"

# push your branch
git push -u origin feat/my-change
```

Keep `CHANGELOG.md` updated; releases derive notes from it.

---

## Build a release

1) **Compile the plugin package** (frontend + backend + zip):

```bash
./create_release.sh
```

Output: `grafana-catalyst-datasource-<version>.zip` in the repo root.  
- If a Git **tag** (e.g., `v1.0.1`) exists, the script uses it as the version.  
- Otherwise it uses the version from `package.json`.

2) **Create an annotated tag for this release** on the current commit:

```bash
git tag -a v1.0.1 -m "Release v1.0.1"
```

3) **Publish to GitHub** (push `main` + the tag, and extract release notes from `CHANGELOG.md`):

```bash
./create-git-release.sh
```

The script will:
- Push `main` and the new tag
- Generate `RELEASE_NOTES.md` from the `CHANGELOG.md` section for that tag

Then finalize the GitHub release in the UI, attaching the generated zip if desired.

---

## Troubleshooting

- Plugin not visible?
  - Confirm Grafana **12.1+**
  - Check container logs for plugin load errors
  - Ensure the plugin directory has the built `dist/` and `plugin.json` (the release script patches `%VERSION%`/`%TODAY%`)
- Backend didn’t start?
  - Release script builds Linux/amd64 binary at `dist/grafana-catalyst-datasource_linux_amd64`

---

## License

Apache-2.0 © extkljajicm

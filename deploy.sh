#!/bin/sh
# Auto-update script run by the updater container (Settings → "Update now").
# Pulls the latest main, rebuilds the app images, and restarts only the app
# services — the updater itself is deliberately not rebuilt so the deploy
# survives, and the database container is never touched.
# The repo is fetched over HTTPS so a PUBLIC repo needs no credentials.
set -e
cd "${REPO_PATH:-/repo}"
# The updater runs as root while the checkout may be owned by the server's
# deploy user; allow git to operate on it regardless of ownership.
git config --global safe.directory "${REPO_PATH:-/repo}"
REPO_URL="${REPO_URL:-https://github.com/cakeru/autostock.git}"
echo "[deploy] fetching ${REPO_URL}"
git fetch "$REPO_URL" main
git reset --hard FETCH_HEAD

COMPOSE_FILE="${REPO_PATH:-/repo}/docker-compose.yml"
LAST_SHA="$(cat "${REPO_PATH:-/repo}/.deploy-sha" 2>/dev/null || true)"
NEW_SHA="$(git rev-parse FETCH_HEAD)"

# Rebuild only the image(s) whose sources actually changed between deploys —
# a backend-only update skips the frontend npm/tsc/vite build entirely. The
# first deploy (no marker) builds both. Compose-file changes rebuild both.
changed() {
  [ -z "$LAST_SHA" ] && return 0
  local out
  out="$(git diff --name-only "$LAST_SHA" "$NEW_SHA" -- "$@" 2>/dev/null)" || return 0
  [ -n "$out" ]
}

BUILD_BACKEND=""
if changed backend docker-compose.yml; then BUILD_BACKEND=backend; fi
BUILD_FRONTEND=""
if changed frontend docker-compose.yml; then BUILD_FRONTEND=frontend; fi

echo "[deploy] rebuilding: ${BUILD_BACKEND:-none} ${BUILD_FRONTEND:-none}"
if [ -n "$BUILD_BACKEND" ] && [ -n "$BUILD_FRONTEND" ]; then
  docker compose -f "$COMPOSE_FILE" build backend &
  B_PID=$!
  docker compose -f "$COMPOSE_FILE" build frontend &
  F_PID=$!
  wait "$B_PID" "$F_PID"
elif [ -n "$BUILD_BACKEND" ]; then
  docker compose -f "$COMPOSE_FILE" build backend
elif [ -n "$BUILD_FRONTEND" ]; then
  docker compose -f "$COMPOSE_FILE" build frontend
fi

docker compose -f "$COMPOSE_FILE" up -d $BUILD_BACKEND $BUILD_FRONTEND
echo "$NEW_SHA" > "${REPO_PATH:-/repo}/.deploy-sha"
echo "[deploy] done"

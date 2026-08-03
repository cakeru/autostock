#!/bin/sh
# Auto-update script run by the updater container (Settings → "Update now").
# Pulls the latest main, rebuilds the app images, and restarts only the app
# services — the updater itself is deliberately not rebuilt so the deploy
# survives, and the database container is never touched.
# The repo is fetched over HTTPS so a PUBLIC repo needs no credentials.
set -e
cd /repo
REPO_URL="${REPO_URL:-https://github.com/cakeru/autostock.git}"
echo "[deploy] fetching ${REPO_URL}"
git fetch "$REPO_URL" main
git reset --hard FETCH_HEAD
echo "[deploy] rebuilding backend + frontend"
docker compose -f /repo/docker-compose.yml up -d --build backend frontend
echo "[deploy] done"

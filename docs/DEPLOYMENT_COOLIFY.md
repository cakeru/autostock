# Deploying AutoStock with Coolify

A step-by-step runbook for a single-shop production deploy on one VPS, using
[Coolify](https://coolify.io) for git-driven deploys, TLS, and env management.

> This is the recommended path. The older `DEPLOYMENT.md` describes generic
> manual options and a different (separate reverse-proxy) architecture — ignore
> it if you're following this.

---

## 0. How the app is wired (read this first)

Three containers, deployed together as **one Coolify "Docker Compose" resource**:

```
              Coolify Traefik (TLS, :443)
                        │
                        ▼
   frontend (nginx :80) ── proxies /api and /uploads ──▶ backend (Go :8080)
                                                              │
                                                              ▼
                                                        postgres :5432
```

The frontend's nginx proxies `/api/` and `/uploads/` to the backend **by service
name** (`http://backend:8080`). That only works if all three services share one
Docker network — which is why they must be deployed as a single compose stack,
**not** as three separate Coolify apps. Only `frontend` is exposed; backend and
Postgres stay internal.

Key facts:
- **Migrations run automatically** on every deploy (the backend image runs
  `migrate ... up` before starting). Keep **one** backend replica so they never race.
- The backend **panics on boot** if `JWT_SECRET` is unset/default in production.
- Migration `000001` seeds an **`admin` / `admin123`** account — you will change
  this immediately after first login (Step 9).

---

## 1. Provision the VPS

- **Spec:** 4 GB RAM / 2 vCPU / ~80 GB NVMe. (The app is tiny; the RAM is for
  Coolify + image builds, which spike during deploy.)
- **Provider/region:** closest to the shop. For Cambodia, choose a **Singapore**
  region (Hetzner Singapore recommended; Contabo/OVH Singapore are alternatives).
- **OS:** Ubuntu 24.04 LTS.
- **Access:** add your SSH public key; disable password login.
- **Firewall:** allow only `22` (SSH), `80`, `443`. Do **not** open Postgres.

```bash
# On the VPS, once:
ufw allow 22 && ufw allow 80 && ufw allow 443 && ufw enable
```

## 2. Point DNS at the VPS

Create an **A record** for your domain (e.g. `autostock.chousour.com`) → the
VPS's public IP. Let it propagate before requesting TLS in Step 7.

## 3. Install Coolify

```bash
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
```

Then open `http://<VPS_IP>:8000`, create the admin account, and (recommended)
put Coolify's own dashboard behind a subdomain with TLS.

## 4. (Recommended) Object storage for images

Put customer/vehicle photos in a bucket so the VPS is stateless for images and
they survive a rebuild.

1. Create a bucket (Cloudflare R2, Backblaze B2, S3, etc.) and an access key.
2. Give the bucket a **public** base URL (R2 public bucket URL or a CDN domain).
3. You'll set `STORAGE_DRIVER` + the `R2_*`/`S3_*` vars in Step 6.

You can start on `local` and switch later — there's a one-time backfill tool
(see [`reference: storage`](#appendix-switching-image-storage-later)).

## 5. Create the resource in Coolify

1. **+ New** → **Project** (e.g. "AutoStock") → add an **Environment**
   (`production`).
2. **+ New Resource** → **Docker Compose** → connect this **Git repository**
   (add a deploy key if private) → branch `main`.
3. Set the **Compose file path** to `docker-compose.coolify.yml`.
4. **Domains:** map your domain to the **`frontend`** service, port **80**.
   Coolify wires Traefik to it.

## 6. Environment variables

In the resource's **Environment Variables**, add the values from
[`.env.production.example`](../.env.production.example). At minimum:

```
DB_PASSWORD=<openssl rand -base64 24>
JWT_SECRET=<openssl rand -base64 48>
APP_ENV=production
LOG_LEVEL=info
STORAGE_DRIVER=r2            # or s3, or local
# ...and the matching R2_* / S3_* values if not local
```

Mark `DB_PASSWORD` and `JWT_SECRET` as **secret**. `DATABASE_URL` is built from
`DB_PASSWORD` inside the compose — don't set it separately. Telegram is
configured **inside the app** later, not here.

## 7. Persistent volumes + TLS

- **Volumes:** confirm `postgres_data` (and `uploads_data` if `STORAGE_DRIVER=local`)
  are **persistent** in Coolify. This is the #1 footgun — an ephemeral
  `postgres_data` wipes the database on every redeploy.
- **TLS:** with DNS resolving (Step 2) and the domain mapped (Step 5), enable
  **Let's Encrypt** for the frontend domain. Coolify obtains and renews the cert.

## 8. First deploy

Click **Deploy**. Coolify builds the images on the box (this is the RAM-hungry
step — normal on 4 GB) and starts the stack. On boot the backend runs the
migrations automatically. Verify:

```bash
# Backend health (from the VPS):
docker ps                              # all three services healthy
curl -fsS http://localhost:8080/health # or check via Coolify's logs

# Then open https://your-domain in a browser — the login screen should load.
```

## 9. First login — SECURE THE ADMIN ACCOUNT

1. Log in with **`admin` / `admin123`** (seeded by migration).
2. **Immediately change the admin password** (Settings/Users → change password).
   Leaving `admin123` in place is an open door.
3. Create real staff accounts with least-privilege roles (mechanics only need
   the low-privilege `install:scan` permission, etc.).
4. Configure the shop: **Settings** → shop name/address/phone, exchange rate,
   tax, payment methods, service-reminder intervals.
5. (Optional) **Settings → Telegram**: add a bot token + chat IDs and route the
   topics you want, including the **Documents** channel for "Send to Telegram".

## 10. Backups (do at least two of these)

The database is the only irreplaceable state (images live in the bucket if you
did Step 4).

- **Provider snapshots** — enable automated daily VPS snapshots at the provider.
  Simplest safety net.
- **In-app monthly backup** — Settings → Telegram, route the **Monthly database
  backup** topic to a channel; the app sends a gzipped `pg_dump` monthly.
  Off-site copy for free.
- **Scheduled dump to the bucket (recommended)** — add a Coolify **Scheduled
  Task** on this resource, container `postgres`, e.g. daily:
  ```bash
  pg_dump -U autostock autostock | gzip > /tmp/autostock-$(date +\%F).sql.gz
  # then push /tmp/*.sql.gz to your bucket (aws/rclone), or just keep the app's
  # Telegram monthly backup + provider snapshots if you prefer fewer moving parts.
  ```

## 11. Updates / redeploys

- Push to `main` → Coolify redeploys (enable **auto-deploy on push** if you want
  it automatic). Migrations run on the new backend's boot.
- **Never run migrations by hand.** If a deploy fails mid-migration, the schema
  version is marked *dirty* and the backend won't start — see Troubleshooting.

## 12. Hardening checklist

- [ ] SSH key-only, password login disabled, firewall = 22/80/443 only.
- [ ] `JWT_SECRET` and `DB_PASSWORD` are strong and unique (not dev values).
- [ ] Admin `admin123` password changed; staff on least-privilege roles.
- [ ] Postgres not exposed to the internet (this compose keeps it internal).
- [ ] TLS enabled and auto-renewing.
- [ ] At least two backup paths (Step 10) and a **tested restore**.
- [ ] Images on R2/S3 (Step 4), or the `uploads_data` volume is included in backups.

---

## Troubleshooting

**Backend won't start / "JWT_SECRET must be set…"** — set a real `JWT_SECRET`
env var and redeploy.

**Database empty after a redeploy** — `postgres_data` wasn't persistent. Fix the
volume in Coolify and restore from a backup.

**Backend crash-loops with a "dirty" migration** — a migration failed partway.
Inspect: `docker exec -it <postgres> psql -U autostock -c "SELECT * FROM schema_migrations;"`.
Resolve the failed statement, set the version clean (`UPDATE schema_migrations
SET dirty=false`), fix the migration, and redeploy. Do **not** patch the schema
by hand and leave migrations out of sync.

**Out-of-memory during build** — you're on <4 GB. Add swap, or build images in CI
and have Coolify pull them instead of building on the box.

**Images 404 after switching to R2/S3** — old rows still point at `/uploads/...`.
Run the backfill (below).

---

## Appendix: switching image storage later

If you started on `local` and later move to a bucket, migrate the existing files
once (the tool ships in the backend image):

```bash
# Set STORAGE_DRIVER=r2|s3 + the keys first, then, from the backend container:
docker exec -it <backend> /app/migrate-storage -dry-run   # preview
docker exec -it <backend> /app/migrate-storage            # move + rewrite URLs
```

Idempotent (only touches URLs still on the local prefix); safe to re-run.

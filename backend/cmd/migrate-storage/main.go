// Command migrate-storage is a one-time backfill that moves already-uploaded
// files off the local disk and into the configured object bucket, rewriting the
// URLs stored in the database as it goes.
//
// Run it with the TARGET bucket driver configured (STORAGE_DRIVER=s3 or r2 and
// the matching keys), while UPLOAD_DIR and STORAGE_PUBLIC_URL still point at the
// existing local files it reads from. It only touches rows whose URL still uses
// the local prefix, so it is safe to re-run — already-migrated rows are skipped.
//
//	# preview
//	docker compose run --rm --entrypoint /app/migrate-storage backend -dry-run
//	# do it
//	docker compose run --rm --entrypoint /app/migrate-storage backend
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/config"
	"github.com/cakeru/autostock/internal/storage"
)

// Every column that holds an uploaded-file URL.
var targets = []struct{ table, col string }{
	{"products", "image_url"},
	{"branches", "logo_url"},
	{"vehicle_photos", "url"},
	{"vehicle_record_photos", "url"},
	{"wheel_service_photos", "url"},
}

func main() {
	dryRun := flag.Bool("dry-run", false, "list what would move without uploading or writing")
	flag.Parse()

	cfg := config.Load()
	if cfg.Storage.Driver == "" || cfg.Storage.Driver == "local" {
		log.Fatal("STORAGE_DRIVER must be a bucket driver (s3 or r2) — there's nothing to migrate to")
	}

	store, err := storage.New(cfg.Storage)
	if err != nil {
		log.Fatalf("build target storage: %v", err)
	}

	oldPrefix := strings.TrimRight(cfg.Storage.PublicURL, "/") // e.g. "/uploads"
	if oldPrefix == "" {
		log.Fatal("STORAGE_PUBLIC_URL (the old local prefix) must be set to find files to migrate")
	}
	srcDir := cfg.Storage.UploadDir

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if *dryRun {
		log.Printf("DRY RUN — reading %q, would upload via %q driver, no changes written", srcDir, cfg.Storage.Driver)
	} else {
		log.Printf("migrating files from %q to the %q bucket…", srcDir, cfg.Storage.Driver)
	}

	var moved, missing, failed int
	for _, t := range targets {
		type job struct {
			id  int64
			url string
		}
		var jobs []job

		rows, err := pool.Query(ctx,
			fmt.Sprintf("SELECT id, %s FROM %s WHERE %s LIKE $1", t.col, t.table, t.col),
			oldPrefix+"/%")
		if err != nil {
			log.Fatalf("query %s.%s: %v", t.table, t.col, err)
		}
		for rows.Next() {
			var j job
			if err := rows.Scan(&j.id, &j.url); err != nil {
				rows.Close()
				log.Fatalf("scan %s: %v", t.table, err)
			}
			jobs = append(jobs, j)
		}
		rows.Close()

		var tableMoved int
		for _, j := range jobs {
			key := strings.TrimPrefix(j.url, oldPrefix+"/")
			path := filepath.Join(srcDir, filepath.Clean("/"+key))

			f, err := os.Open(path)
			if err != nil {
				log.Printf("  ! %s#%d: file missing on disk (%s) — skipping", t.table, j.id, key)
				missing++
				continue
			}

			if *dryRun {
				f.Close()
				log.Printf("  · %s#%d would move %s", t.table, j.id, key)
				moved++
				tableMoved++
				continue
			}

			ct := mime.TypeByExtension(filepath.Ext(key))
			if ct == "" {
				ct = "application/octet-stream"
			}
			newURL, err := store.Save(ctx, key, f, ct)
			f.Close()
			if err != nil {
				log.Printf("  ! %s#%d upload failed: %v", t.table, j.id, err)
				failed++
				continue
			}

			if _, err := pool.Exec(ctx,
				fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id = $2", t.table, t.col),
				newURL, j.id); err != nil {
				log.Printf("  ! %s#%d db update failed: %v", t.table, j.id, err)
				failed++
				continue
			}
			moved++
			tableMoved++
		}
		if len(jobs) > 0 {
			log.Printf("  %s.%s: %d/%d", t.table, t.col, tableMoved, len(jobs))
		}
	}

	log.Printf("done — %d %s, %d missing on disk, %d failed",
		moved, map[bool]string{true: "to move", false: "moved"}[*dryRun], missing, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

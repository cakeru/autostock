// updater is the tiny "update now" agent: it holds the Docker socket and the
// repo checkout, and on POST /deploy runs the deploy script in the background
// (this container is never rebuilt by the deploy, so the process survives).
// The main backend proxies to it so the button stays behind the app's auth.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

func repoPath() string {
	if p := os.Getenv("REPO_PATH"); p != "" {
		return p
	}
	return "/repo"
}

var (
	mu      sync.Mutex
	running bool
	lastRun time.Time
)

func main() {
	http.HandleFunc("POST /deploy", handleDeploy)
	http.HandleFunc("GET /status", handleStatus)
	port := os.Getenv("UPDATER_PORT")
	if port == "" {
		port = "8081"
	}
	fmt.Printf("updater listening on :%s\n", port)
	http.ListenAndServe(":"+port, nil)
}

func handleDeploy(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	if running {
		writeJSON(w, http.StatusConflict, map[string]any{"status": "deploying", "message": "An update is already running"})
		return
	}
	repo := repoPath()
	if _, err := os.Stat(repo + "/deploy.sh"); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "error", "message": "deploy.sh not found in the repo checkout"})
		return
	}

	running = true
	lastRun = time.Now()
	go func() {
		defer func() { mu.Lock(); running = false; mu.Unlock() }()
		f, err := os.OpenFile(repo+"/deploy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return
		}
		defer f.Close()
		ts := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(f, "\n===== deploy started %s =====\n", ts)
		cmd := exec.Command("/bin/sh", repo+"/deploy.sh")
		cmd.Stdout = f
		cmd.Stderr = f
		_ = cmd.Run()
		fmt.Fprintf(f, "===== deploy finished %s (exit %v) =====\n", time.Now().Format("2006-01-02 15:04:05"), cmd.ProcessState)
	}()

	writeJSON(w, http.StatusAccepted, map[string]any{"data": map[string]any{"status": "deploying", "message": "Update started — the app will restart shortly"}})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	tail := ""
	if b, err := os.ReadFile(repoPath() + "/deploy.log"); err == nil {
		if n := len(b); n > 2000 {
			b = b[n-2000:]
		}
		tail = string(b)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{
		"status":   map[bool]string{true: "deploying", false: "idle"}[running],
		"last_run": lastRun.Format(time.RFC3339),
		"log_tail": tail,
	}})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

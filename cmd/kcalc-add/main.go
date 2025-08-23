package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ssmythe/kcalc-add/handler"
)

// These get overridden at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown" // ISO-8601 UTC
	BuiltBy = "local"
)

// accessLog is tiny middleware to log method, path, and duration.
func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	// Route logs to a file if LOG_FILE is set; otherwise stderr.
	if p := os.Getenv("LOG_FILE"); p != "" {
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("failed to open LOG_FILE %q: %v", p, err)
		} else {
			log.SetOutput(f)
			defer f.Close()
		}
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("kcalc-add %s\ncommit: %s\nbuilt:  %s\nby:     %s\n", Version, Commit, Date, BuiltBy)
		return
	}

	// Handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/add", handler.AddHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Printf("healthz write error: %v", err)
		}
	})
	handlerWithLog := accessLog(mux)

	// Server (note: no Addr yet; set it only in fixed-port path)
	srv := &http.Server{
		Handler:           handlerWithLog,
		ReadHeaderTimeout: 5 * time.Second,
	}

	port := getenv("PORT", "8080")
	portFile := os.Getenv("PORT_FILE")

	// Startup banner (helpful in logs/artifacts)
	log.Printf("starting kcalc-add %s (commit %s, built %s) PORT=%s", Version, Commit, Date, port)

	// Start server
	go func() {
		// Ephemeral port: PORT=0 -> OS picks, we write it to PORT_FILE (if set)
		if port == "0" {
			ln, err := net.Listen("tcp", ":0")
			if err != nil {
				log.Fatalf("listen :0: %v", err)
			}
			actual := ln.Addr().(*net.TCPAddr).Port
			log.Printf("Listening on :%d (ephemeral)", actual)
			if portFile != "" {
				_ = os.WriteFile(portFile, []byte(fmt.Sprintf("%d", actual)), 0644)
			}
			if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server error: %v", err)
			}
			return
		}

		// Fixed port path
		srv.Addr = ":" + port
		log.Printf("Listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

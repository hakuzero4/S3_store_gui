package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/hakup/s3store/internal/api"
	"github.com/hakup/s3store/internal/config"
	"github.com/hakup/s3store/internal/static"
)

var version = "0.1.0"

func main() {
	addr := flag.String("addr", "", "listen address (default from config / env S3STORE_ADDR)")
	configPath := flag.String("config", "", "config.json path (default: next to executable, or env S3STORE_CONFIG)")
	noBrowser := flag.Bool("no-browser", false, "do not open browser")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	cfgPath := firstNonEmpty(*configPath, os.Getenv("S3STORE_CONFIG"))
	if cfgPath == "" {
		var err error
		cfgPath, err = config.DefaultPath()
		if err != nil {
			log.Fatalf("config path: %v", err)
		}
	}

	store, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	listen := firstNonEmpty(*addr, os.Getenv("S3STORE_ADDR"), store.ListenAddr)
	if listen == "" {
		listen = "127.0.0.1:17890"
	}

	apiServer := api.New(store)
	mux := http.NewServeMux()
	mux.Handle("/api/", apiServer.Handler())
	mux.Handle("/", static.Handler())

	srv := &http.Server{
		Addr:              listen,
		Handler:           withLog(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("listen %s: %v", listen, err)
	}
	actual := ln.Addr().String()
	url := "http://" + actual
	if host, port, err := net.SplitHostPort(actual); err == nil && (host == "127.0.0.1" || host == "::1") {
		url = "http://127.0.0.1:" + port
	}

	log.Printf("S3 Store v%s listening on %s", version, url)
	log.Printf("config: %s", cfgPath)

	open := store.OpenBrowser && !*noBrowser && os.Getenv("S3STORE_NO_BROWSER") == ""
	if open {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = openBrowser(url)
		}()
	}

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func withLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, code: 200}
		next.ServeHTTP(ww, r)
		if stringsHasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s %d %s", r.Method, r.URL.Path, ww.code, time.Since(start).Truncate(time.Millisecond))
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func stringsHasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

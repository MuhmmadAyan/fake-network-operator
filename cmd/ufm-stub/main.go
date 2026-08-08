package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	listenAddr string
	setupLog   = ctrl.Log.WithName("setup")
)

func init() {
	flag.StringVar(&listenAddr, "listen-addr", ":8443", "Address to listen on")
}

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("starting ufm-stub", "addr", listenAddr)

	mux := http.NewServeMux()

	pkeys := map[string]string{
		"0x7fff": "default",
	}

	mux.HandleFunc("/ufmRest/resources/pkeys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pkeys)
	})

	mux.HandleFunc("/ufmRest/resources/ports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ports": []string{}})
	})

	mux.HandleFunc("/ufmRest/app/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			setupLog.Error(err, "failed to start HTTP server")
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	setupLog.Info("shutting down ufm-stub")
	server.Shutdown(context.Background())
}

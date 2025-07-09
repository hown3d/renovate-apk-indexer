package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hown3d/renovate-apk-indexer/pkg/apk"
	"github.com/hown3d/renovate-apk-indexer/pkg/renovate"
)

const wolfiIndex = "https://packages.wolfi.dev/os/x86_64/APKINDEX.tar.gz"

var (
	updateInterval = flag.Int("update-interval", 4, "update interval of the apk package index in hours")
	apkIndexUrls   = flag.String("apk-index-url", wolfiIndex, "comma-separated URLs of the apk indexes to get the package information from")
	logLevel       = new(slog.Level)
	logOutput      = flag.String("log-output", "text", "representation for logs (text,json)")
)

func main() {
	flag.TextVar(logLevel, "log-level", slog.LevelInfo, "log level")
	flag.Parse()
	var (
		l       *slog.Logger
		logOpts = &slog.HandlerOptions{
			Level: logLevel,
		}
	)

	switch *logOutput {
	case "json":
		jsonHandler := slog.NewJSONHandler(os.Stdout, logOpts)
		l = slog.New(jsonHandler)
	case "text":
		textHandler := slog.NewTextHandler(os.Stdout, logOpts)
		l = slog.New(textHandler)
	}
	slog.SetDefault(l)

	urls := strings.Split(*apkIndexUrls, ",")

	apkContext := apk.New(http.DefaultClient, urls)
	slog.Info("retrieving apk packages", "urls", urls)
	apkPackages, err := apkContext.GetApkPackages()
	if err != nil {
		slog.Error("error getting apk packages", "err", err)
	}

	ticker := time.NewTicker(time.Duration(*updateInterval) * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				slog.Info("updating apk packages")
				newPackages, err := apkContext.GetApkPackages()
				if err != nil {
					slog.Error("error updating apk packages", "err", err)
					continue
				}
				apkPackages = newPackages
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", healthHandler)
	mux.HandleFunc("/livez", healthHandler)
	mux.HandleFunc("/{package}", func(w http.ResponseWriter, r *http.Request) {
		packageName := r.PathValue("package")
		packages, ok := apkPackages[packageName]
		if !ok {
			slog.Debug("package not found", "packageName", packageName)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%s not found in apkIndex", packageName)
			return
		}

		slog.Debug("packages found", "packageName", packageName)
		datasource := renovate.TransformAPKPackage(packages)
		if err := json.NewEncoder(w).Encode(datasource); err != nil {
			slog.Error("encoding datasource", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "internal server error")
		}
	})

	slog.Info("serving on :3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		slog.Error("serving http ", "err", err)
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("ok"))
}

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/PolRuff/urlshort/api"
	"github.com/PolRuff/urlshort/internal/audit"
	"github.com/PolRuff/urlshort/internal/config"
	"github.com/PolRuff/urlshort/internal/handler"
	"github.com/PolRuff/urlshort/internal/middleware"
	"github.com/PolRuff/urlshort/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	_ "net/http/pprof"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Вывод информации о сборке
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	var repo repository.Repository

	if cfg.DatabaseDsn != "" {
		repo, err = repository.NewSQLRepository(cfg.DatabaseDsn)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open sql repository")
		}
	} else if cfg.FileStoragePath != "" {
		// If a file path is provided, create a FileRepository
		repo, err = repository.NewFileRepository(cfg.FileStoragePath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize file repository")
		}
	} else {
		// Otherwise, create an in-memory repository
		repo = repository.NewMemoryRepository()
	}

	defer repo.Close()

	var auditSinks []audit.Sink
	if cfg.AuditFile != "" {
		auditSinks = append(auditSinks, audit.NewFileSink(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		auditSinks = append(auditSinks, audit.NewHTTPSink(cfg.AuditURL))
	}
	auditManager := audit.NewManager(auditSinks...)

	var trustedNet *net.IPNet

	if cfg.TrustedSubnet != "" {
		// Parse trusted subnet CIDR
		_, trustedNet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			log.Fatal().Err(err).Msg("Invalid trusted subnet configuration")
		}
	}

	h := handler.New(repo, cfg.BaseURL, cfg.SecretKey, auditManager, trustedNet)

	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.Logger)

	r.Post("/", h.ShortenHandler)
	r.Post("/api/shorten", h.ShortenAPIHandler)
	r.Post("/api/shorten/batch", h.ShortenBatchAPIHandler)
	r.Delete("/api/user/urls", h.DeleteUserUrlsHandler)
	r.Get("/api/user/urls", h.UserUrlsHandler)
	r.Get("/api/internal/stats", h.StatsHandler)
	r.Get("/{id}", h.RedirectHandler)
	r.Get("/ping", h.PingHandler)

	log.Debug().Msgf("Base URL for short links: %s", cfg.BaseURL)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		const pprofPort = ":9090"
		log.Debug().Msgf("pprof server is running on http://localhost%s/debug/pprof/", pprofPort)

		if err := http.ListenAndServe(pprofPort, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("pprof server failed")
		}
	}()

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	wg.Add(1)
	go func() {
		defer wg.Done()

		var serverErr error

		if cfg.EnableHTTPS {
			log.Debug().Msgf("Server is running on https://%s", cfg.ServerAddr)
			serverErr = server.ListenAndServeTLS("server.crt", "server.key")
		} else {
			log.Debug().Msgf("Server is running on http://%s", cfg.ServerAddr)
			serverErr = server.ListenAndServe()
		}

		if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
			log.Error().Err(serverErr)
		}
	}()

	// gRPC сервер
	grpcServer := grpc.NewServer()
	grpcHandler := handler.NewGRPCHandler(repo, auditManager, cfg.BaseURL, cfg.SecretKey)
	api.RegisterShortenerServiceServer(grpcServer, grpcHandler)

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Debug().Msgf("gRPC server is running on %s", cfg.GRPCServerAddr)
		lis, err := net.Listen("tcp", cfg.GRPCServerAddr)
		if err != nil {
			log.Error().Err(err).Msg("failed to listen")
			return
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Error().Err(err).Msg("gRPC server failed")
		}
	}()

	// ждём завершения процедуры graceful shutdown
	<-ctx.Done()
	log.Info().Msg("Received shutdown signal, gracefully shutting down...")

	// Graceful shutdown с таймаутом 30 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		log.Debug().Msg("Server exited gracefully")
	}

	// Завершение gRPC сервера
	grpcServer.GracefulStop()

	// Дожидаемся завершения горутин
	wg.Wait()
}

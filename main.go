package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Resavin/pr-reviewers/internal/config"
	"github.com/Resavin/pr-reviewers/internal/controller"
	"github.com/Resavin/pr-reviewers/internal/generated"
	"github.com/Resavin/pr-reviewers/internal/repository"
	"github.com/Resavin/pr-reviewers/internal/service"
)

//go:embed api/openapi.yml
var openapiSpec []byte

//go:embed static/docs.html
var swaggerHTML []byte

func main() {
	cfg := config.Load()

	log.Printf("DB URL: %s\n", cfg.DBURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Successfully connected to database")

	teamRepo := repository.NewTeamRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	prRepo := repository.NewPullRequestRepository(pool)

	userSvc := service.NewUserService(userRepo)
	// maybe not so good
	prSvc := service.NewPullRequestService(prRepo, userRepo)
	teamSvc := service.NewTeamService(teamRepo, userSvc, prSvc)

	apiServer := controller.NewServer(teamSvc, userSvc, prSvc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	generated.HandlerFromMux(apiServer, r)

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		_, _ = w.Write(openapiSpec)
	})

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(swaggerHTML)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on :%s\n", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited properly")
}

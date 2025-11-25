package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pr-reviewer-service/internal/config"
	"pr-reviewer-service/internal/database"
	"pr-reviewer-service/internal/handler"
	"pr-reviewer-service/internal/repository"
	"pr-reviewer-service/internal/service"
)

func main() {
	cfg := config.Load()

	log.Println("Connection to database...")
	var db *sql.DB
	var err error
	maxRetries := 5
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = database.Connect(cfg)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)
	statisticsService := service.NewStatisticsService(prRepo)

	teamHandler := handler.NewTeamHandler(teamService)
	userHandler := handler.NewUserHandler(userService)
	prHandler := handler.NewPRHandler(prService)
	statisticsHandler := handler.NewStatisticsHandler(statisticsService)

	mux := http.NewServeMux()

	mux.HandleFunc("/team/add", teamHandler.CreateTeam)
	mux.HandleFunc("/team/get", teamHandler.GetTeam)

	mux.HandleFunc("/users/setIsActive", userHandler.SetIsActive)
	mux.HandleFunc("/users/getReview", userHandler.GetUserReviews)

	mux.HandleFunc("/pullRequest/create", prHandler.CreatePR)
	mux.HandleFunc("/pullRequest/merge", prHandler.MergePR)
	mux.HandleFunc("/pullRequest/reassign", prHandler.ReassignReviewer)

	mux.HandleFunc("/statistics", statisticsHandler.GetStatistics)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting server on port %s...", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited properly")
}
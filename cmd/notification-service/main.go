package main

import (
	"booking-system/configs"
	"booking-system/internal/notification/ws"
	"booking-system/pkg/kafka"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/middleware"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := configs.LoadConfig()

	if cfg.Mode == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	hub := ws.NewHub()

	// Create worker context
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		hub.Run(workerCtx)
	}()

	// inital kafka reader for notification topic
	kafkaReader := kafka.NewReader(
		cfg.Kafka.Brokers,
		constants.NotificationTopic,
		constants.NotificationServiceWSPollGroup,
	)
	defer kafkaReader.Close()

	wsConsumer := ws.NewWsConsumer(kafkaReader, hub)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := wsConsumer.Consume(workerCtx); err != nil && err != context.Canceled {
			log.Printf("WS consumer: stopped with error: %v", err)
		}
	}()

	r := gin.Default()

	r.Use(middleware.WsCorsMiddleware())

	wsHanler := ws.NewHandler(hub)

	r.GET("/ws/buses/:id", wsHanler.Handle)

	srv := &http.Server{
		Addr:    ":" + "8082",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	// 1. Shutdown HTTP server first, allowing 5 seconds to finish active requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown failed: %v", err)
	} else {
		log.Println("HTTP server shut down successfully")
	}

	// 2. Stop workers
	cancelWorkers()

	// Wait for all background workers to stop (with a timeout of 5 seconds)
	workersDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workersDone)
	}()

	select {
	case <-workersDone:
		log.Println("All background workers stopped successfully")
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for background workers to stop")
	}

}

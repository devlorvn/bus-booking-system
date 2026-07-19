package main

import (
	"booking-system/configs"
	"booking-system/internal/notification/ws"
	"booking-system/pkg/redis"
	"booking-system/pkg/shared/constants"
	"booking-system/pkg/shared/events"
	"booking-system/pkg/shared/metrics"
	"booking-system/pkg/shared/middleware"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := configs.LoadConfig()

	if cfg.Mode == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	hub := ws.NewHub()
	redisClient := redis.NewClient(&cfg.Redis)
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	var wg sync.WaitGroup

	// init worker for broadcast message
	wg.Add(1)
	go func() {
		defer wg.Done()
		hub.Run(workerCtx)
	}()

	// register pub/sub
	wg.Add(1)
	go func() {
		defer wg.Done()

		pubsub := redisClient.Subscribe(workerCtx, constants.WsChanel)
		defer pubsub.Close()

		ch := pubsub.Channel()
		slog.Info("Websocket service: Subcribe to channel %s", slog.String("channel", constants.WsChanel))

		for {
			select {
			case <-workerCtx.Done():
				slog.Info("Websocket service: Worker stopped")
				return
			case msg, ok := <-ch:
				if !ok {
					slog.Error("Websocket service: Channel closed")
					return
				}
				var event events.KafkaWsEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					slog.Error("Websocket service: Failed to unmarshal event: %v", slog.String("error", err.Error()))
					continue
				}
				// hub.BroadcastMessage(event)

				var data map[string]interface{}

				if err := json.Unmarshal(event.Data, &data); err != nil {
					slog.Error("Websocket service: Failed to unmarshal event data: %v", slog.String("error", err.Error()))
					continue
				}

				hub.Broadcast(event.BusID, ws.BroadcastMessage{
					Event: constants.EventType(event.Event),
					Data:  data,
				})
			}
		}
	}()

	r := gin.Default()
	r.Use(metrics.GinMetricsMiddleware())
	r.Use(middleware.WsCorsMiddleware())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	wsHandler := ws.NewHandler(hub)
	r.GET("/ws/buses/:id", wsHandler.Handle)

	srv := &http.Server{
		Addr:    ":8082",
		Handler: r,
	}

	go func() {
		slog.Info("Websocket service: Started at %s", slog.String("addr", srv.Addr))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Websocket service: Failed to start server: %v", slog.String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	slog.Info("Websocket service: Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Websocket service: Server forced to shutdown: %v", slog.String("error", err.Error()))
	}

	cancelWorkers()
	wg.Wait()

	slog.Info("Websocket service: Server exited")
}

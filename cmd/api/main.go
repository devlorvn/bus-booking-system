package main

import (
	"booking-system/configs"
	"booking-system/pkg/postgres"
	"booking-system/pkg/redis"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := configs.LoadConfig()
	_, err := postgres.NewPostgres(&cfg.Database)
	if err != nil {
		panic(err)
	}

	_ = redis.NewClient(&cfg.Redis)

	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"data": "pong!!!",
		})
	})

	router.Run(":" + cfg.Port)
}

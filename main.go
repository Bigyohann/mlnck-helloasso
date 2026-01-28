package main

import (
	"time"

	"mlnck/api"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	limit "github.com/yangxikun/gin-limit-by-key"
	"golang.org/x/time/rate"
)

func main() {
	_ = godotenv.Load(".env" + ".local")
	_ = godotenv.Load()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://mlnck.fr"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(limit.NewRateLimiter(func(c *gin.Context) string {
		return c.ClientIP() // limit rate by client ip
	}, func(c *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(
			rate.Every(100*time.Millisecond),
			10,
		), time.Hour
	}, func(c *gin.Context) {
		c.AbortWithStatus(429)
	}))

	r.GET("/forms", api.GetFormsHandler)
	_ = r.Run()
}

package api

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	limit "github.com/yangxikun/gin-limit-by-key"
	"golang.org/x/time/rate"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "" {
		allowedOrigins = []string{"*"}
	}

	for _, origin := range allowedOrigins {
		log.Println("Allowed origin:", origin)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Allow cross-origin resource sharing for images and other resources
	r.Use(func(c *gin.Context) {
		c.Header("Cross-Origin-Resource-Policy", "cross-origin")
		c.Next()
	})

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

	r.GET("/forms", GetFormsHandler)
	r.GET("/proxy-image", ProxyImageHandler)

	return r
}

package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggerMiddleware(log *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		c.Next()
		log.Infof("| %d | %s | %s | %s |",
			c.Writer.Status(),
			time.Since(now),
			c.Request.Method,
			c.Request.URL.String(),
		)
	}
}

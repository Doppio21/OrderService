package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"orderservice/internal/orderdb"
	"orderservice/internal/schema"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	shutdownTimeout = 5 * time.Second
)

type Config struct {
	Address string
}

type Dependencies struct {
	Log *logrus.Logger
	DB  orderdb.OrderDB
}

type Server struct {
	cfg  Config
	deps Dependencies

	log *logrus.Entry
}

func NewServer(cfg Config, deps Dependencies) *Server {
	return &Server{
		cfg:  cfg,
		deps: deps,
		log:  deps.Log.WithField("component", "server"),
	}
}

func (s *Server) Run(ctx context.Context) error {
	router := gin.New()
	router.Use(LoggerMiddleware(s.log))

	router.GET("/", s.uiHandler)
	router.GET("orders/:id", s.getHandler)
	router.GET("orders/", s.listHandler)

	var (
		srv = &http.Server{
			Addr:    s.cfg.Address,
			Handler: router,
		}
	)

	serverClosed := make(chan struct{})
	go func() {
		s.log.Info("server started")
		defer close(serverClosed)
		if err := srv.ListenAndServe(); err == nil && err != http.ErrServerClosed {
			s.log.Fatalf("listen and serve: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down server gracefully")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	case <-serverClosed:
	}

	s.log.Info("server finished")
	return nil
}

func (s *Server) getHandler(c *gin.Context) {
	id := c.Param("id")
	res, err := s.deps.DB.GetOrder(c, schema.OrderUID(id))
	if s.replyError(c, err) {
		return
	}

	c.JSON(http.StatusOK, &res)
}

func (s *Server) listHandler(c *gin.Context) {
	res, err := s.deps.DB.ListOrders(c)
	if s.replyError(c, err) {
		return
	}

	c.JSON(http.StatusOK, &res)
}

func (s *Server) uiHandler(c *gin.Context) {
	t, err := template.ParseFiles("ui/templates/order.html")
	if s.replyError(c, err) {
		return
	}

	c.Status(200)
	err = t.ExecuteTemplate(c.Writer, "order.html", nil)
	if s.replyError(c, err) {
		return
	}
}

func (s *Server) replyError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	resp := ErrorResponse{Message: err.Error()}
	code := http.StatusInternalServerError

	switch {
	case errors.Is(err, orderdb.ErrNotFound):
		code = http.StatusNotFound
	}

	c.JSON(code, &resp)
	return true
}

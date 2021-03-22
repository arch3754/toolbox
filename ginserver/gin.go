package ginserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/arch3754/toolbox/log"
)

type GinServer struct {
	s      *http.Server
	Engine *gin.Engine
}

func NewGinServer(ginMode, addr string, readTimeout, writeTimeout time.Duration) *GinServer {
	gin.SetMode(ginMode)
	return &GinServer{
		s: &http.Server{
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			MaxHeaderBytes: 1 << 20,
			Addr:           addr,
		},
		Engine: gin.New(),
	}
}
func (g *GinServer) ConfigMiddleware(middleware ...gin.HandlerFunc) {
	g.Engine.Use(middleware...)
}

func (g *GinServer) Start() {
	go func() {
		log.Rlog.Info("starting http server, listening on:", g.s.Addr)
		g.s.Handler = g.Engine
		if err := g.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Rlog.Critical("listening %s occur error: %s\n", g.s.Addr, err)
		}
	}()
}

// Shutdown http server
func (g *GinServer) Close() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := g.s.Shutdown(ctx); err != nil {
		log.Rlog.Critical("cannot shutdown http server:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Rlog.Info("shutdown http server timeout of 5 seconds.")
	default:
		log.Rlog.Info("http server stopped")
	}
}

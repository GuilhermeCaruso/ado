package server

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	addr  string
	webFS embed.FS
}

func New(addr string, webFS embed.FS) *Server {
	return &Server{addr: addr, webFS: webFS}
}

func (s *Server) Start() error {
	sub, err := fs.Sub(s.webFS, "web")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

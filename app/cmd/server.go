package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serveHttp() error {
	//serve http - addr, routes,
	srv := http.Server{
		Addr:     fmt.Sprintf(":%d", app.config.HttpPort),
		Handler:  app.routes(),
		ErrorLog: app.errorLog,
	}

	//graceful shutdown
	shutDownErrChan := make(chan error)
	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutDownErrChan <- srv.Shutdown(ctx)
	}()

	//starting server
	app.infoLog.Printf("Starting server on port %d", app.config.HttpPort)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	//stopping server
	app.errorLog.Printf("Server stopped")
	app.wg.Wait()

	return nil
}

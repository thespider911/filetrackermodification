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

	// done channel to signal when all cleanup is complete
	done := make(chan bool)

	// add context that we can cancel  with
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start a goroutine for graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		app.infoLog.Println("Server is shutting down...")

		// cancel the context to signal all goroutines to stop
		cancel()

		// attempt to shut down the server gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			app.errorLog.Printf("Server shutdown error: %v", err)
		}

		// stop the service (this should stop worker and timer threads)
		app.stopService()

		close(done)
	}()

	app.infoLog.Printf("Starting server on port %d", app.config.HttpPort)

	// start the service (this should start worker and timer threads)
	app.startService()

	// start the server
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// wait for done signal
	<-done
	app.infoLog.Println("Server stopped")

	return nil
}

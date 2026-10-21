package api

import (
	"errors"
	"net/http"
	"time"
)

func Start(handler http.Handler, address string) (*http.Server, <-chan error) {
	server := &http.Server{
		Addr:        address,
		Handler:     handler,
		ReadTimeout: 5 * time.Second,
	}
	done := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		done <- err
	}()

	return server, done
}

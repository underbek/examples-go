package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func listen(ctx context.Context) error {
	srv := http.Server{
		Addr:        ":8080",
		Handler:     nil,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(srv.ListenAndServe)

	eg.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			fmt.Println("Shutdown error:", err)
		}
		return err
	})

	return eg.Wait()
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		fmt.Println("Done")
		w.Write([]byte("done\n"))
	})

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return listen(ctx)
	})

	fmt.Println(eg.Wait())
}

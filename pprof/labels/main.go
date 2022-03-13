package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	rp "runtime/pprof"
	"strings"
)

const (
	addr = ":8080" // адрес сервера
)

func getID(uri string) string {
	uris := strings.Split(uri, "/")
	return uris[len(uris)-1]
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	var slice []int

	for i := 0; i < 10_000_000; i++ {
		slice = append(slice, i)
	}

	if _, err := w.Write([]byte(getID(r.RequestURI))); err != nil {
		panic(err)
	}
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		labels := rp.Labels("handler", getID(r.RequestURI))

		ctx := rp.WithLabels(r.Context(), labels)

		rp.Do(ctx, labels, func(ctx context.Context) {
			next(w, r)
		})
	}
}

func main() {
	http.HandleFunc("/", middleware(profileHandler))
	fmt.Println(http.ListenAndServe(addr, nil)) // запускаем сервер
}

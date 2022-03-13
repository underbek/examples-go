package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"runtime/trace"
)

const (
	addr = ":8080" // адрес сервера
)

func profileHandler(w http.ResponseWriter, r *http.Request) {
	ctx, task := trace.NewTask(r.Context(), "profileHandler")
	defer task.End()

	reg := trace.StartRegion(ctx, "createSlice")
	var slice []int
	reg.End()

	reg = trace.StartRegion(ctx, "growSlice")
	for i := 0; i < 10_000_000; i++ {
		slice = append(slice, i)
	}
	reg.End()

	reg = trace.StartRegion(ctx, "writeResponse")
	if _, err := w.Write([]byte("done")); err != nil {
		panic(err)
	}
	reg.End()
}

func main() {
	http.HandleFunc("/profile", profileHandler)
	fmt.Println(http.ListenAndServe(addr, nil)) // запускаем сервер
}

package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof" // подключаем пакет pprof
	"runtime"
	"sync"
	"time"
)

const (
	addr = ":8081" // адрес сервера
)

var (
	slice []int
	mtx   sync.Mutex
)

func add() {
	i := 0
	for {
		i++
		mtx.Lock()
		slice = append(slice, i)
		mtx.Unlock()
		time.Sleep(time.Millisecond)
	}
}

func get() {
	for {
		mtx.Lock()
		if len(slice) > 10 {
			slice = slice[1:]
		}
		mtx.Unlock()
		time.Sleep(time.Millisecond)
	}
}

func test() {
	c := make(chan bool)
	<-c
}

func main() {
	// выключены по дефолту
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	// будем отдавать профиль только внутренним пользователям
	debugMux := http.NewServeMux()
	// только runtime pprof
	debugMux.HandleFunc("/debug/pprof/", pprof.Index)

	debugMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debugMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debugMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go add() // запускаем полезную нагрузку в фоне
	go add() // запускаем полезную нагрузку в фоне
	go get() // запускаем полезную нагрузку в фоне
	go get() // запускаем полезную нагрузку в фоне
	go test()

	// можно выставить debug=2 для goroutine (можно видеть статус горутин)
	fmt.Println(http.ListenAndServe(addr, debugMux)) // запускаем сервер
}

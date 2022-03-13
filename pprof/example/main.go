package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
)

const (
	addr    = ":8080"    // адрес сервера
	maxSize = 10_000_000 // будем растить слайс до 10 миллионов элементов
)

func foo() {
	// наша полезная нагрузка
	for {
		var s []int
		for i := 0; i < maxSize; i++ {
			s = append(s, i)
		}
	}
}

func main() {
	go foo()                                    // запускаем полезную нагрузку в фоне
	fmt.Println(http.ListenAndServe(addr, nil)) // запускаем сервер
}

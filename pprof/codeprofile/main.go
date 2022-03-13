package main

import (
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
)

func foo() {
	var s []int
	for i := 0; i < 100_000_000; i++ {
		// поменяем тут
		if i == 90_000_000 {
			// создаём файл журнала профилирования памяти
			fmem, err := os.Create(`pprof/codeprofile/mem.profile`)
			if err != nil {
				panic(err)
			}
			defer fmem.Close()
			runtime.GC() // получаем статистику по использованию памяти
			if err := pprof.WriteHeapProfile(fmem); err != nil {
				panic(err)
			}
		}
		s = append(s, i)
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// создаём файл журнала профилирования cpu
	fcpu, err := os.Create(`pprof/codeprofile/cpu.profile`)
	if err != nil {
		panic(err)
	}
	defer fcpu.Close()
	if err := pprof.StartCPUProfile(fcpu); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	foo()

	if _, err := w.Write([]byte("done")); err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/profile", profileHandler)
	http.ListenAndServe(":8080", nil)
}

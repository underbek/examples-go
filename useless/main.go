package main

import (
	"fmt"
	"runtime"
)

func main() {
	source := []int{1, 2, 3, 4, 5}

	target := make([]int, len(source))
	copy(target, source)

	fmt.Println(target)
	fmt.Println(runtime.NumCPU())
}

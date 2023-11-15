package main

import (
	"csbbrokerpakaws/acceptance-tests/helpers/brokerpaks"
	"fmt"
)

func main() {
	for _, v := range brokerpaks.Versions() {
		fmt.Println(v)
	}
}

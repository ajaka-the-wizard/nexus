package main

import "gateway/internal"

func main() {
	if err := internal.Listen(); err != nil {
		panic(err)
	}
}

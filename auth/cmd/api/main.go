package main

import "auth/internal"

func main() {
	if err := internal.Listen(); err != nil {
		panic(err)
	}
}

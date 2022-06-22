package main

import (
	"math/rand"
	"time"

	"github.com/vitalvas/secretshare/backend"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	backend.Execute()
}

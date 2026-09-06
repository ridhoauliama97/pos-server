package main

import (
	"github.com/ridhoauliama97/pos-server/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}

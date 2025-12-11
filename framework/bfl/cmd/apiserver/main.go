package main

import (
	"bytetrade.io/web3os/bfl/cmd/apiserver/app"
)

func main() {
	cmd := app.NewAPPServerCommand()

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

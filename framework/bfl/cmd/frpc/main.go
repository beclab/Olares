package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"bytetrade.io/web3os/bfl/internal/frpc/agent"
)

func main() {
	config := flag.String("config", "/etc/frpc/frpc.toml", "Path to the binding's FRPC configuration")
	healthAddress := flag.String("health-address", ":7401", "Health server listen address")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := agent.Run(ctx, *config, *healthAddress); err != nil {
		log.Fatal(err)
	}
}

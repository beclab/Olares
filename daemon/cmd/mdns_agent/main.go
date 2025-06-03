package main

import (
	"os"
	"os/signal"
	"syscall"

	"bytetrade.io/web3os/terminusd/internel/mdns"
	"k8s.io/klog/v2"
)

func main() {
	port := 18088
	s, err := mdns.NewServer(port)
	if err != nil {
		klog.Error(err)
	}

	defer s.Close()
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit

	klog.Info("agent is quit")
}

/*
 * Copyright 2020 kingSoft cloud.
 *
 * @Author : zhangliang7@kingsoft.com
 * @Date   : 2021/9/3 15:01
 * @Desc   :
 *
 */

package signal

import (
	"os"
	"os/signal"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func StopCh() chan struct{} {
	var c = make(chan struct{})

	sc := make(chan os.Signal, 2)
	signal.Notify(sc, shutdownSignals...)

	go func() {
		<-sc
		c <- struct{}{}

		<-sc
		os.Exit(1) // second signal. Exit directly
	}()
	return c
}

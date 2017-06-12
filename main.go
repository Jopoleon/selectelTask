package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Jopoleon/selectelTask/app"
	"github.com/fclairamb/ftpserver/server"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	ftpServer *server.FtpServer
)

func main() {
	flag.Parse()
	ftpServer = server.NewFtpServer(app.NewSampleDriver())

	go signalHandler2()

	err := ftpServer.ListenAndServe()
	if err != nil {
		log15.Error("Problem listening", "err", err)
	}
}

func signalHandler2() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM)
	for {
		switch <-ch {
		case syscall.SIGTERM:
			ftpServer.Stop()
			break
		}
	}
}

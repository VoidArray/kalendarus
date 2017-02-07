package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("kalendarus %s\n", Version)
		os.Exit(0)
	}
	if err := initConfig(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("Starting kalendarus")
	logrus.Infof("Configuration: %#v", config)

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	processor := NewProcessor(config, stopChan, doneChan, errChan, messenger, backend)

	go processor.Process()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			logrus.Error(err)
		case s := <-signalChan:
			logrus.Infof("Captured %v. Exiting...", s)
			close(doneChan)
		case <-doneChan:
			if err := processor.SaveState(); err != nil {
				logrus.Fatal(err)
			}
			os.Exit(0)
		}
	}
}

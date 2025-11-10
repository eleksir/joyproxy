package joyproxy

import (
	"joyproxy/internal/log"
	"syscall"
)

// SigHandler handles signals from OS.
func SigHandler() {
	for {
		var s = <-SigChan
		switch s {
		case syscall.SIGINT:
			log.Info("Got SIGINT, quitting")
		case syscall.SIGTERM:
			log.Info("Got SIGTERM, quitting")
		case syscall.SIGQUIT:
			log.Info("Got SIGQUIT, quitting")
		case syscall.SIGHUP:
			log.Info("Got SIGHUP, reopening logs")
			log.ReOpenLog()

			continue

		// Make new iteration since we 've got signal we not interested in.
		default:
			continue
		}

		Shutdown = true

		if err := Server.Close(); err != nil {
			log.Errorf("Unable to close server: %s", err.Error())
		}

		log.Close()

		QuitChan <- true

		break
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

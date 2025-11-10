// Package main represents entry point function of application.
package main

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"joyproxy/internal/joyproxy"
	"joyproxy/internal/log"
)

// main application entry point.
func main() {
	var (
		err     error
		logfile *os.File
	)

	joyproxy.LoadConf()

	if joyproxy.Cfg.Logfile != "" {
		logfile, err = os.OpenFile(joyproxy.Cfg.Logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Unable to open %s: %s", joyproxy.Cfg.Logfile, err)
		}
	} else {
		logfile = os.Stderr
	}

	log.Init(joyproxy.Cfg.Loglevel, logfile)

	log.Info("Install handler to /joyproxy/ location")
	http.HandleFunc("/joyproxy/", joyproxy.JoyproxyHandler)

	log.Info("Install handler to /joyurl location")
	http.HandleFunc("/joyurl", joyproxy.JoyurlHandler)

	// It's time to set up signal trapper.
	signal.Notify(joyproxy.SigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)

	go joyproxy.SigHandler()

	log.Infof("Listening on port %v\n", joyproxy.Cfg.Port)

	// To prevent slow loris ddos, we have to define custom handler because builtin http.ListenAndServe() have no
	// timeout and sits here forever.
	joyproxy.Server = &http.Server{
		Addr:              joyproxy.Cfg.Port,
		ReadHeaderTimeout: time.Duration(joyproxy.Cfg.Timeout) * time.Second,
		ReadTimeout:       time.Duration(joyproxy.Cfg.Timeout) * time.Second,
		WriteTimeout:      time.Duration(joyproxy.Cfg.Timeout) * time.Second,
		IdleTimeout:       time.Duration(joyproxy.Cfg.Timeout) * time.Second,
		ErrorLog:          &log.Logger,
	}

	for {
		if joyproxy.Shutdown {
			time.Sleep(100 * time.Microsecond)

			continue
		}

		err = joyproxy.Server.ListenAndServe()

		if errors.Is(err, http.ErrServerClosed) {
			log.Info("Shutdown joyproxy")

			break
		}

		log.Errorf("Listener error: %s", err)
	}

	<-joyproxy.QuitChan
	os.Exit(0)
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

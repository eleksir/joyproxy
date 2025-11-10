package joyproxy

// Config represents application config data structure.
type (
	Config struct {
		Port     string
		Proto    string
		Loglevel string
		Logfile  string
		Timeout  int
	}
)

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

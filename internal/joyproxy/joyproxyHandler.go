package joyproxy

import (
	"joyproxy/internal/log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// JoyproxyHandler request handler for /joyproxy location. It reads data from reactor.cc and immediately streams it to
// client.
func JoyproxyHandler(w http.ResponseWriter, r *http.Request) { //nolint: revive
	var err error

	if r.Method != http.MethodGet {
		log.Infof("Request method is not GET, but %s, sending 404 to client", r.Method)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	dlReq := r
	joyPath := strings.Split(r.RequestURI, "/")
	// joyPath[0] is empty, joyPath[1] is joyproxy.
	joyHost := joyPath[2]

	if !regexp.MustCompile("^img[0-9]+[.]reactor[.]cc$").MatchString(joyHost) {
		log.Infof("Supplied proxy target host is not match with ^img[0-9]+[.]reactor[.]cc$ pattern: %s, sening 404 to client", joyHost)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	joyPath = joyPath[3:]
	dlReq.RequestURI = "/" + strings.Join(joyPath, "/")

	if !regexp.MustCompile("^/pics/post/mp4/.+[.]mp4$").MatchString(dlReq.RequestURI) {
		log.Infof("Supplied uri is not match with /pics/post/mp4/.+[.]mp4$ pattern: %s, sending 404 to client", dlReq.RequestURI)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	dlReq.Host = joyHost
	dlReq.URL, err = url.Parse(dlReq.RequestURI)

	if err != nil {
		log.Infof("Unable to parse %s via url.Parse(): %s, sending 404 to client", dlReq.RequestURI, err)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	// Forge request as it comes from browser… kinda.
	if dlReq.Header.Get("Referer") != "" {
		dlReq.Header.Del("Referer")
	}

	log.Debugf("Replacing/appending Referer header value with https://old.reactor.cc/all")
	dlReq.Header.Add("Referer", "https://old.reactor.cc/all")

	if dlReq.Header.Get("User-Agent") != "" {
		dlReq.Header.Del("User-Agent")
	}

	ua := userAgentString[rand.Intn(len(userAgentString)-1)]
	log.Debugf("Replacing User-Agent header value with %s", ua)
	dlReq.Header.Add("User-Agent", ua)

	if dlReq.Header.Get("Accept") != "" {
		dlReq.Header.Del("Accept")
	}

	dlReq.Header.Add("Accept", "*/*")

	if dlReq.Header.Get("Accept-Encoding") != "" {
		dlReq.Header.Del("Accept-Encoding")
	}

	dlReq.Header.Add("Accept-Encoding", "identity;q=1, *;q=0")

	if dlReq.Header.Get("Accept-Language") != "" {
		dlReq.Header.Del("Accept-Language")
	}

	dlReq.Header.Add("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")

	if dlReq.Header.Get("Range") != "" {
		dlReq.Header.Del("Range")
	}

	log.Debugf("Replacing/appending Range header value with bytes=0-")
	dlReq.Header.Add("Range", "bytes=0-")
	log.Debug(spew.Sdump(dlReq))

	u := "https://" + joyHost
	target, err := url.Parse(u)

	if err != nil {
		log.Infof("Unable to parse %s via url.Parse(): %s, sending 404 to client", u, err)
		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write([]byte("404 Not Found")); err != nil {
			log.Infof("Unable to write response to client: %s", err)
		}

		return
	}

	joyproxy := httputil.NewSingleHostReverseProxy(target)

	joyproxy.Transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Duration(Cfg.Timeout) * time.Second,
			KeepAlive: time.Duration(Cfg.Timeout) * time.Second,
		}).Dial,
		TLSHandshakeTimeout: time.Duration(Cfg.Timeout) * time.Second,
		DisableKeepAlives:   true,
		DisableCompression:  true,
	}

	joyproxy.ServeHTTP(w, dlReq)
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

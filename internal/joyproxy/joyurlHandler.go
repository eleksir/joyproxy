package joyproxy

import (
	"fmt"
	"html"
	"net/http"
	"regexp"

	"joyproxy/internal/log"

	"github.com/davecgh/go-spew/spew"
)

// JoyurlHandler request handler function for /joyurl location.
func JoyurlHandler(w http.ResponseWriter, r *http.Request) {
	var (
		param          = r.URL.Query()
		proxyURLString = " "
	)

	log.Debugf("Parse url into chunks: %s", spew.Sdump(param))

	re := regexp.MustCompile("^https?://img[0-9]+[.]reactor[.]cc/.+[.][Ww][Ee][Bb][Mm]$")

	if len(param) > 0 && re.MatchString(param["joyurl"][0]) {
		log.Debug("Url match with ^https?://img[0-9]+[.]reactor[.]cc/.+[.][Ww][Ee][Bb][Mm]$ pattern")
		// https://img1.reactor.cc/pics/post/webm/видосик.webm

		p := regexp.MustCompile("/").Split(param["joyurl"][0], -1)
		log.Debugf("Split url string by / delimiter: %s", spew.Sdump(p))

		if len(p) > 6 && p[5] == "webm" {
			log.Debug("Substring webm is detected in url, let's make joyproxy url")

			file := p[6][:len(p[6])-5]

			proxyURLString = fmt.Sprintf(
				"%s://%s/joyproxy/%s/%s/%s/mp4/%s.mp4",
				Cfg.Proto,
				r.Host,
				p[2],
				p[3],
				p[4],
				file,
			)

			proxyURLString = html.EscapeString(proxyURLString)
		} else {
			log.Debug("Substring webm not detected in url, skip making joyproxy url")
		}
	} else {
		log.Info("Url is not match with ^https?://img[0-9]+[.]reactor[.]cc/.+[.][Ww][Ee][Bb][Mm]$ regex, skipping")
	}

	htmlText := fmt.Sprintf(postForm, proxyURLString)

	if _, err := fmt.Fprintln(w, htmlText); err != nil {
		log.Infof("Unable to write response to client: %s", err)
	}
}

/* vim: set ft=go noet ai ts=4 sw=4 sts=4: */

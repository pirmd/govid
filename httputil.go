package main

import (
	"log"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/realclientip/realclientip-go"
)

// noopHandler is an http Handler that does nothing.
func noopHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// loggingHandler provides a middleware that logs Request metrics in the Apache
// Common log format (see http://httpd.apache.org/docs/2.2/logs.html#common).
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)

		log.Printf("%s - %s [%s] \"%s %s %s\" %d %d",
			getClientIP(r),
			GetUsernameFromContext(r),
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.String(),
			r.Proto,
			m.Code,
			m.Written,
		)
	})
}

// getClientIP tries its best to find client's real IP.
// see also: https://adam-p.ca/blog/2022/03/x-forwarded-for/
func getClientIP(r *http.Request) string {
	strat := realclientip.NewChainStrategy(
		realclientip.Must(realclientip.NewSingleIPHeaderStrategy("X-Real-IP")),
		realclientip.Must(realclientip.NewRightmostNonPrivateStrategy("X-Forwarded-For")),
		realclientip.RemoteAddrStrategy{},
	)

	return strat.ClientIP(r.Header, r.RemoteAddr)
}

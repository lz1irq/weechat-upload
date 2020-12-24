package main

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type LoggableResponse struct {
	http.ResponseWriter
	statusCode int
}

func (l *LoggableResponse) WriteHeader(status int) {
	l.statusCode = status
	l.ResponseWriter.WriteHeader(status)
}

func makeRequestLog(r *http.Request) log.Fields {
	return log.Fields{
		"source": r.RemoteAddr,
		"method": r.Method,
		"host":   r.Host,
		"URI":    r.RequestURI,
	}
}

func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestInfo := makeRequestLog(r)

		respWriter := &LoggableResponse{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(respWriter, r)
		requestInfo["response"] = respWriter.statusCode

		log.WithFields(requestInfo).Info("request")
	})
}

func noIndex(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			log.WithFields(makeRequestLog(r)).Error("Missing basic auth")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if user != conf.AuthUsername || pass != conf.AuthPassword {
			log.WithFields(makeRequestLog(r)).Error("Wrong username/password")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

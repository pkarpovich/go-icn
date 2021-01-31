package main

import (
	"errors"
	"golang.org/x/net/http2"
	"net/http"
	"strconv"
)

type ServerHttp2 struct {
	server      *http.Server
	certificate string
	key         string
}

type ServerHttp2Config struct {
	Port             int
	Certificate      string
	Key              string
	StaticFolderName string
	Listeners        map[string]func(w http.ResponseWriter, r *http.Request)
}

func CreateHttp2Server(cfg ServerHttp2Config) (s ServerHttp2, err error) {
	if cfg.Port == 0 {
		err = errors.New("port must be non-zero")
		return
	}

	if cfg.Certificate == "" {
		err = errors.New("certificate must be specified")
		return
	}

	if cfg.Key == "" {
		err = errors.New("key must be specified")
		return
	}

	s.server = &http.Server{
		Addr: ":" + strconv.Itoa(cfg.Port),
	}

	s.certificate = cfg.Certificate
	s.key = cfg.Key

	err = http2.ConfigureServer(s.server, nil)

	http.Handle("/", http.FileServer(http.Dir(cfg.StaticFolderName)))

	for path, handler := range cfg.Listeners {
		http.HandleFunc(path, handler)
	}

	return s, err
}

func (s *ServerHttp2) Listen() (err error) {
	err = s.server.ListenAndServeTLS(
		s.certificate, s.key)
	if err != nil {
		err = errors.New("failed during server listen and serve")
		return
	}

	return
}

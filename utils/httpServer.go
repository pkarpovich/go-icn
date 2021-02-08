package utils

import (
	"errors"
	"net/http"
	"strconv"
)

type ServerHttp struct {
	server      *http.Server
}

type ServerHttpConfig struct {
	Port             int
	StaticFolderName string
	Listeners        map[string]func(w http.ResponseWriter, r *http.Request)
}

func CreateHttpServer(cfg ServerHttpConfig) (s ServerHttp, err error) {
	if cfg.Port == 0 {
		err = errors.New("port must be non-zero")
		return
	}

	s.server = &http.Server{
		Addr: ":" + strconv.Itoa(cfg.Port),
	}

	if cfg.StaticFolderName != "" {
		http.Handle("/", http.FileServer(http.Dir(cfg.StaticFolderName)))
	}

	for path, handler := range cfg.Listeners {
		http.HandleFunc(path, handler)
	}

	return s, err
}

func (s *ServerHttp) Listen() (err error) {

	err = s.server.ListenAndServe()
	if err != nil {
		err = errors.New("failed during server listen and serve")
		return
	}

	return
}

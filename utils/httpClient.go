package utils

import (
	"net/http"
)

type ClientHttp struct {
	Client  *http.Client
}

type ClientHttpConfig struct {
}

func CreateHttpClient(cfg ClientHttpConfig) (c ClientHttp, err error) {

	c.Client = &http.Client{
		Transport: &http.Transport{},
	}

	return c, err
}

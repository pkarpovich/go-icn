package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pkarpovich/go-icn/utils"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

type controllerContext struct {
	providers map[string]bool
	redis     *utils.RedisClient
}

func main() {
	httpPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		log.Fatalf("HTTP port is not a number")
	}
	if httpPort == 0 {
		log.Fatalf("HTTP port is not provided")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatalf("Redis address is not provided")
	}

	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("Redis DB is not a number")
	}
	if redisDb == 0 {
		log.Fatalf("Redis address is not provided")
	}

	cc := &controllerContext{
		providers: map[string]bool{},
		redis: utils.CreateRedisClient(context.Background(), redisAddr, "", redisDb),
	}

	httpServer, err := utils.CreateHttpServer(utils.ServerHttpConfig{
		Port: httpPort,
		Listeners: map[string]func(w http.ResponseWriter, r *http.Request){
			"/query":      cc.handleQuery(),
			"/connect":    cc.handleProviderConnect(),
			"/disconnect": cc.handleProviderDisconnect(),
			"/providers":  cc.handleProvidersList(),
		},
	})
	if err != nil {
		log.Fatalf("can't create http2 server, err: %e", err)
	}

	err = httpServer.Listen()

}

func (cc *controllerContext) handleQuery() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := r.URL.Query().Get("fileName")

		if value, err := cc.redis.Get(fileName); err == nil {
			fmt.Println("from cache")
			fmt.Fprintf(w, value)
			return
		}

		httpClient, err := utils.CreateHttpClient(utils.ClientHttpConfig{})
		if err != nil {
			http.Error(w, "can't create http client", http.StatusInternalServerError)
			return
		}

		urls := make(chan string, len(cc.providers))

		for providerIp := range cc.providers {
			providerIp := providerIp
			go func() {
				fileUrl := fmt.Sprintf("http://%s/query?fileName=%s", providerIp, fileName)
				providerUrl := fmt.Sprintf("%s&checkOnly=true", fileUrl)
				resp, _ := httpClient.Client.Get(providerUrl)
				if resp != nil && resp.StatusCode == http.StatusOK {
					urls <- fileUrl
				} else {
					urls <- ""
				}
			}()
		}

		index := 0
		for url := range urls {
			index++
			if url != "" {
				cc.redis.Set(fileName, url)
				fmt.Fprintf(w, url)
				return
			}

			if index == len(cc.providers) {
				break
			}
		}

		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (cc *controllerContext) handleProviderConnect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		providerAddr, err := getUserIp(r.RemoteAddr)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		cc.providers[providerAddr] = true

		w.WriteHeader(200)
	}
}

func (cc *controllerContext) handleProviderDisconnect() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		providerAddr, err := getUserIp(r.RemoteAddr)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		delete(cc.providers, providerAddr)
		w.WriteHeader(200)
	}
}

func (cc *controllerContext) handleProvidersList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(cc.providers)
		if err != nil {
			log.Fatalf("can't marshal providers list, err: %e", err)
		}
	}
}

func getUserIp(remoteAddr string) (string, error) {
	ip, _, err := net.SplitHostPort(remoteAddr)

	if err != nil {
		return "", errors.New(fmt.Sprintf("userip: %q is not IP:port", remoteAddr))
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		return "", errors.New(fmt.Sprintf("userip: %q is not IP:port", remoteAddr))
	}

	return fmt.Sprintf("%s:%v", ip, 8080), err
}

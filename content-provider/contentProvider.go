package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pkarpovich/go-icn/utils"
	"github.com/tebeka/atexit"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type controllerContext struct {
	redisClient *utils.RedisClient
	contentFolder string
}

func main() {
	httpPort, err := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if err != nil {
		log.Fatalf("HTTP port is not a number")
	}
	if httpPort == 0 {
		log.Fatalf("HTTP port is not provided")
	}

	dirName := os.Getenv("CONTENT_DIR_NAME")
	if dirName == "" {
		log.Fatalf("Content directory path is not provided")
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
		log.Fatalf("Redis DB is not provided")
	}

	managerAddr := os.Getenv("MANAGER_URL")
	if managerAddr == "" {
		log.Fatalf("Certificate path key is not provided")
	}

	cc := &controllerContext{
		redisClient:   utils.CreateRedisClient(context.Background(), redisAddr, "", redisDb),
		contentFolder: dirName,
	}

	err = connectToManager(managerAddr)
	if err != nil {
		log.Fatalf("can't connect to manager, err: %e", err)
	}
	atexit.Register(disconnectFromManager(managerAddr))

	httpServer, err := utils.CreateHttpServer(utils.ServerHttpConfig{
		Port:        httpPort,
		Listeners: map[string]func(w http.ResponseWriter, r *http.Request){
			"/query": cc.handleQuery(),
		},
	})
	if err != nil {
		log.Fatalf("can't create http2 server, err: %e", err)
	}

	_, err = indexContent(dirName, cc.redisClient)
	if err != nil {
		log.Fatalf("can't indexing files, err: %e", err)
	}

	err = httpServer.Listen()
	if err != nil {
		log.Fatalf("error during listen, err: %e", err)
	}

	atexit.Exit(0)
}

func indexContent(dirName string, redisClient *utils.RedisClient) ([]string, error) {
	var dirFiles []string

	err := filepath.Walk(dirName, visit(&dirFiles))
	if err != nil {
		return nil, err
	}

	for _, fileName := range dirFiles {
		err := redisClient.Set(fileName, fileName)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Can't save file name (%s) in cache, error: %e", fileName, err))
		}
	}

	return dirFiles, nil
}

func visit(dirFiles *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		*dirFiles = append(*dirFiles, info.Name())

		return nil
	}
}

func connectToManager(managerAddr string) error {
	client, err := utils.CreateHttpClient(utils.ClientHttpConfig{})
	if err != nil {
		return err
	}

	_, err = client.Client.Get(fmt.Sprintf("%s/connect", managerAddr))
	if err != nil {
		return err
	}

	return nil
}

func disconnectFromManager(managerAddr string) func() {
	return func() {
		client, _ := utils.CreateHttpClient(utils.ClientHttpConfig{})
		client.Client.Get(fmt.Sprintf("%s/disconnect", managerAddr))
	}
}

func (cc *controllerContext) handleQuery() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := r.URL.Query().Get("fileName")
		checkOnly := r.URL.Query().Get("checkOnly")

		value, err := cc.redisClient.Get(fileName)
		if err != nil {
			http.Error(w, fmt.Sprintf("can't find file name: %s, err: %e", fileName, err), http.StatusNotFound)
			return
		}

		if checkOnly != "" {
			err = json.NewEncoder(w).Encode(value)
			if err != nil {
				http.Error(w, fmt.Sprintf("can't marshal redis result, err: %e", err), http.StatusInternalServerError)
			}
			return
		}

		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", cc.contentFolder, fileName))
		if err != nil {
			http.Error(w, "can't read requested file", http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, fileName, time.Now(),   bytes.NewReader(data))
	}
}

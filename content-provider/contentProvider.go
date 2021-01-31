package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	dirName := os.Getenv("CONTENT_DIR_NAME")
	if dirName == "" {
		log.Fatalf("Content directory path not provide")
	}

	redisClient := CreateRedisClient(context.Background(), "localhost:6379", "", 0)
	httpServer, err := CreateHttp2Server(ServerHttp2Config{
		Port:        8080,
		Certificate: "./localhost.pem",
		Key:         "./localhost-key.pem",
		Listeners: map[string]func(w http.ResponseWriter, r *http.Request){
			"/query": handleQuery,
		},
	})
	if err != nil {
		log.Fatalf("can't create http2 server, err: %e", err)
	}

	indexContent(dirName, redisClient)

	err = httpServer.Listen()
	if err != nil {
		log.Fatalf("error during listen, err: %e", err)
	}
}

func indexContent(dirName string, redisClient *RedisClient) []string {
	var dirFiles []string

	err := filepath.Walk(dirName, visit(&dirFiles))
	if err != nil {
		log.Fatal(err)
	}

	for _, fileName := range dirFiles {
		err := redisClient.Set(fileName, fileName)
		if err != nil {
			log.Fatalf("Can't save file name (%s) in cache, error: %e", fileName, err)
		}
	}

	return dirFiles
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

func handleQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

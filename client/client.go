package main

import (
	"bufio"
	"fmt"
	"github.com/pkarpovich/go-icn/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	managerUrl := os.Getenv("MANAGER_URL")
	if managerUrl == "" {
		log.Fatalf("Manager url is not provided")
	}

	downloadFolderPath := os.Getenv("DOWNLOAD_FOLDER_PATH")
	if downloadFolderPath == "" {
		log.Fatalf("Download folder path is not provided")
	}

	httpClient, err := utils.CreateHttpClient(utils.ClientHttpConfig{})
	if err != nil {
		log.Fatalf("can't create http client err: %e", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		switch input[0] {
		case '>':
			{
				fileName := input[2:]
				fileUrl, err := checkIfFileExist(httpClient, fileName, managerUrl)
				if err != nil {
					log.Printf("can't check file availability, err: %e", err)
					continue
				}

				if fileUrl == "" {
					continue
				}

				err = downloadNewFile(httpClient, fileUrl, downloadFolderPath, fileName)
				if err != nil {
					log.Printf("can't download file, err: %e", err)
					continue
				}

				log.Printf("Downloaded file: %s/%s", downloadFolderPath, fileName)
				break
			}
		default:
			{
				log.Printf("unknown command %s\n", input)
			}

		}
	}
}

func checkIfFileExist(httpClient utils.ClientHttp, fileName, managerUrl string) (string, error) {
	managerQuery := fmt.Sprintf("%s/query?fileName=%s", managerUrl, fileName)

	resp, err := httpClient.Client.Get(managerQuery)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return string(bodyBytes), nil
	}

	return "", nil
}

func downloadNewFile(httpClient utils.ClientHttp, fileUrl, downloadFolderPath, fileName string) error {
	resp, err := httpClient.Client.Get(fileUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fmt.Sprintf("%s/%s", downloadFolderPath, fileName))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

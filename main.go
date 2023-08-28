package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"net/http"
	"os"
	"syscall"
	"time"
)

func main() {
	resp, err := http.Get("https://muxigame.github.io/deploy_shadowsocks/shell.sh")
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	file, err := os.Create("./shell.sh")
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	UpdateConfig()
	for range ticker.C {
		UpdateConfig()
	}
}
func UpdateConfig() {
	config, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/config.json")
	randomConfig, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/random-config.json")
	for _, value := range gjson.ParseBytes(randomConfig).Array() {
		fmt.Println(value.String())
		config, err := sjson.SetBytes(config, value.String(), syscall.GUID{})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		file, err := os.Create("./config.json")
		_, err = file.Write(config)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = file.Close()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
func GetJsonFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	defer func(Body io.ReadCloser) {
		var err = Body.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	//jsonConfig := map[string]interface{}{}

	byte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return byte, nil
}

package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {

	err := DownloadFile("https://muxigame.github.io/deploy_shadowsocks/install_node.sh", "./install_node.sh")
	if err != nil {
		fmt.Printf("Download file error%s", err.Error())
		return
	}
	err = DownloadFile("https://muxigame.github.io/deploy_shadowsocks/restart.sh", "./restart.sh")
	if err != nil {
		fmt.Printf("Download file error%s", err.Error())
		return
	}

	ExecuteCmd("sh", "-c", "chmod +x ./install_node.sh")
	ExecuteCmd("bash", "-c", "./install_node.sh")

	ticker := time.NewTicker(144 * time.Hour)
	defer ticker.Stop()

	UpdateConfig()

	ExecuteCmd("sh", "-c", "chmod +x ./restart.sh")
	ExecuteCmd("bash", "-c", "./restart.sh")

	for range ticker.C {
		UpdateConfig()
	}
}

func UpdateConfig() {
	config, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/config.json")
	randomConfig, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/random-config.json")
	for _, value := range gjson.ParseBytes(randomConfig).Array() {
		fmt.Println(value.String())
		config, err := sjson.SetBytes(config, value.String(), uuid.New().String())
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

	byte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return byte, nil
}

func DownloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func ExecuteCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %s......", err.Error())
		return err
	}

	go func() {
		err := asyncLog(stdout)
		if err != nil {
			log.Printf("Error asyncLog: %s......", err.Error())
		}
	}()
	go func() {
		err := asyncLog(stderr)
		if err != nil {
			log.Printf("Error asyncLog: %s......", err.Error())
		}
	}()

	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command execution: %s......", err.Error())
		return err
	}
	return nil
}

func asyncLog(reader io.ReadCloser) error {
	cache := ""
	buf := make([]byte, 1024, 1024)
	for {
		num, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "closed") {
				err = nil
			}
			return err
		}
		if num > 0 {
			oByte := buf[:num]
			//h.logInfo = append(h.logInfo, oByte...)
			oSlice := strings.Split(string(oByte), "\n")
			line := strings.Join(oSlice[:len(oSlice)-1], "\n")
			fmt.Printf("%s%s\n", cache, line)
			cache = oSlice[len(oSlice)-1]
		}
	}
	return nil
}

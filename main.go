package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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

	fmt.Println("Try install server.....")
	_ = ExecuteCmd("sh", "-c", "chmod +x ./install_node.sh")
	err = ExecuteCmd("bash", "-c", "./install_node.sh")
	if err != nil {
		fmt.Println(" Install server failed.....")
		return
	}

	ticker := time.NewTicker(144 * time.Hour)

	defer ticker.Stop()

	UpdateConfig()
	for range ticker.C {
		UpdateConfig()
	}
}

func UpdateConfig() {
	fmt.Println("Try start server.....")

	byteInstallConfig, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/install_config.json")
	byteVpnServerConfig, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/vpn_server_config.json")
	byteVpnClientConfig, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/vpn_client_config.json")
	byteServerCenterConfig, _ := GetJsonFromUrl("https://muxigame.github.io/deploy_shadowsocks/server_center.json")

	serverCenter := gjson.ParseBytes(byteServerCenterConfig)
	server := serverCenter.Get("server").String()
	port := serverCenter.Get("port").String()
	register := serverCenter.Get("register").String()
	_ = serverCenter.Get("live").String()
	address := server + ":" + port

	installConfig, err := UnmarshalInstallConfig(byteInstallConfig)
	if err != nil {
		fmt.Println("parse install config failed.....")
		return
	}
	ProcessJson(byteVpnServerConfig, byteVpnClientConfig, installConfig)

	_ = ExecuteCmd("sh", "-c", "chmod +x ./restart.sh")
	err = ExecuteCmd("bash", "-c", "./restart.sh")

	if err != nil {
		fmt.Println("Start server failed.....")
		return
	}
	fmt.Println("Start server success.....")

	RegisterConfig(address+register, []byte(gjson.GetBytes(byteVpnServerConfig, "outbounds").Array()[0].String()))
}

func RegisterConfig(url string, body []byte) {
	post, err := http.Post(url, "application/json", bytes.NewReader(body))
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(post.Body)
	if err != nil {
		fmt.Println("register node field....." + err.Error())
		return
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
	fmt.Printf("download %s config success.....\n", url)
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
	cmdReader, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(cmdReader)
	done := make(chan bool)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
		done <- true
	}()
	err := cmd.Start()
	if err != nil {
		log.Printf("Error start: %s......", err.Error())
		return err
	}
	<-done
	err = cmd.Wait()
	if err != nil {
		log.Printf("Error wait: %s......", err.Error())
		return err
	}
	return err
}

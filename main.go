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

	err := DownloadFile("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/install_node.sh", "./install_node.sh")
	if err != nil {
		fmt.Printf("Download install_node.sh file error:%s", err.Error())
		return
	}
	err = DownloadFile("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/restart.sh", "./restart.sh")
	if err != nil {
		fmt.Printf("Download restart.sh file error:%s", err.Error())
		return
	}

	fmt.Println("Try install server.....")

	err = DownloadFile("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/v2ray.key", "/usr/local/etc/v2ray/v2ray.key")
	if err != nil {
		fmt.Printf("Download v2ray.key file error:%s", err.Error())
		return
	}

	err = DownloadFile("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/v2ray.pem", "/usr/local/etc/v2ray/v2ray.pem")
	if err != nil {
		fmt.Printf("Download v2ray.pem file error:%s", err.Error())
		return
	}

	fmt.Println("Save tls file success: /etc/v2ray/v2ray.pem")

	_ = ExecuteCmd("sh", "-c", "chmod +x ./install_node.sh")
	err = ExecuteCmd("bash", "-c", "./install_node.sh")
	if err != nil {
		fmt.Println("Install server failed.....")
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

	byteInstallConfig, _ := GetJsonFromUrl("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/install_config.json")
	byteVpnServerConfig, _ := GetJsonFromUrl("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/vpn_server_config.json")
	byteVpnClientConfig, _ := GetJsonFromUrl("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/vpn_client_config.json")
	byteServerCenterConfig, _ := GetJsonFromUrl("https://Yess-VPN-Pte-Ltd.github.io/deploy_shadowsocks/server_center.json")

	serverCenter := gjson.ParseBytes(byteServerCenterConfig)
	server := serverCenter.Get("server").String()
	port := serverCenter.Get("port").String()
	register := serverCenter.Get("register").String()
	_ = serverCenter.Get("live").String()
	address := server + ":" + port

	installConfig, err := UnmarshalInstallConfig(byteInstallConfig)
	if err != nil {
		fmt.Println("Parse install config failed.....")
		return
	}
	byteVpnServerConfig, byteVpnClientConfig = ProcessJson(byteVpnServerConfig, byteVpnClientConfig, installConfig)

	_ = ExecuteCmd("sh", "-c", "chmod +x ./restart.sh")
	err = ExecuteCmd("bash", "-c", "./restart.sh")

	if err != nil {
		fmt.Println("Start server failed.....")
		return
	}
	fmt.Println("Start server success.....")

	registerConfig := gjson.GetBytes(byteVpnClientConfig, "outbounds").Array()[0].String()

	fmt.Println("register config:\n" + registerConfig)
	RegisterConfig(address+register, []byte(registerConfig))
}

func RegisterConfig(url string, body []byte) {
	post, err := http.Post(url, "application/json", bytes.NewReader(body))
	defer func(Resp *http.Response) {
		if Resp == nil {
			fmt.Println("Response is null: " + err.Error())
			return
		}
		err := Resp.Body.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(post)
	if err != nil {
		fmt.Println("Register node field: " + err.Error())
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

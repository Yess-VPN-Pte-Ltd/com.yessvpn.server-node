package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"math/rand"
	"net"
	"os"
)

func ProcessJson(serverConfig []byte, clientConfig []byte, installConfig InstallConfig) ([]byte, []byte) {
	var err error

	// server config
	for _, random := range installConfig.ServerConfig.Random {
		if random.Type != nil {
			if *random.Type == "guid" {
				serverConfig, err = sjson.SetBytes(serverConfig, random.Path, uuid.New().String())
			}
		} else if random.Set != nil {
			serverConfig, err = sjson.SetBytes(serverConfig, random.Path, random.Set[rand.Intn(len(random.Set))])
		} else if random.Range != nil {
			randomRange := random.Range
			serverConfig, err = sjson.SetBytes(serverConfig, random.Path, rand.Intn(randomRange.Min+randomRange.Max)+randomRange.Min)
		}
		if err != nil {
			fmt.Printf("Set server config %s failed: %s\n", random.Path, err.Error())
			return nil, nil
		}
	}

	if err == nil {
		fmt.Printf("Set server config success...\n")
	}
	// client config
	for _, copyValue := range installConfig.ClientConfig.Copy {
		clientConfig, err = sjson.SetBytes(clientConfig, copyValue.Client, gjson.GetBytes(serverConfig, copyValue.Server).Value())
		if err != nil {
			fmt.Printf("Set client config %s failed: %s\n", copyValue.Client, err.Error())
			return nil, nil
		}
	}

	if err == nil {
		fmt.Printf("Set client config success...\n")
	}

	ip, err := GetLocalIP()
	if err != nil {
		fmt.Printf("Get client ip failed: %s\n", err.Error())
		return nil, nil
	}

	clientConfig, err = sjson.SetBytes(clientConfig, installConfig.ClientConfig.Address, ip)
	if err != nil {
		fmt.Printf("Set client ip failed: %s\n", err.Error())
		return nil, nil
	}

	SaveConfig(installConfig.ClientConfig.Storage, clientConfig)
	SaveConfig(installConfig.ServerConfig.Storage, serverConfig)
	return serverConfig, clientConfig
}

func SaveConfig(path string, data []byte) {
	file, err := os.Create(path)
	_, err = file.Write(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = file.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Save config to %s success...\n", path)
}

func GetLocalIP() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		return ipAddr.IP.String(), nil
	}
	return
}

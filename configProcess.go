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
				serverConfig, _ = sjson.SetBytes(serverConfig, random.Path, uuid.New().String())
			}
		}
		if random.Set != nil {
			serverConfig, _ = sjson.SetBytes(serverConfig, random.Path, random.Set[rand.Intn(len(random.Set))])
		}
		if random.Range != nil {
			randomRange := random.Range
			serverConfig, _ = sjson.SetBytes(serverConfig, random.Path, rand.Intn(randomRange.Min+randomRange.Max)+randomRange.Min)
		}
	}

	// client config
	for _, copyValue := range installConfig.ClientConfig.Copy {
		clientConfig, err = sjson.SetBytes(clientConfig, copyValue.Client, gjson.GetBytes(serverConfig, copyValue.Server))
		if err != nil {
			fmt.Println(err.Error())
			return nil, nil
		}
	}

	ip, err := GetLocalIP()
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil
	}
	clientConfig, err = sjson.SetBytes(clientConfig, installConfig.ClientConfig.Address, ip)
	if err != nil {
		fmt.Println(err.Error())
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

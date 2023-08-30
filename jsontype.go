package main

import "encoding/json"

type Outbounds struct {
	Protocol string   `json:"protocol"`
	Settings Settings `json:"settings"`
}

type Settings struct {
	Servers []Server `json:"servers"`
}

type Server struct {
	Address  string `json:"address"`
	Method   string `json:"method"`
	Ota      bool   `json:"ota"`
	Password string `json:"password"`
	Port     int64  `json:"port"`
}

func UnmarshalInstallConfig(data []byte) (InstallConfig, error) {
	var r InstallConfig
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *InstallConfig) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type InstallConfig struct {
	ServerConfig ServerConfig `json:"serverConfig"`
	ClientConfig ClientConfig `json:"clientConfig"`
}

type ClientConfig struct {
	Storage string `json:"storage"`
	Address string `json:"address"`
	Copy    []Copy `json:"copy"`
}

type Copy struct {
	Server string `json:"server"`
	Client string `json:"client"`
}

type ServerConfig struct {
	Storage string   `json:"storage"`
	Random  []Random `json:"random"`
}

type Random struct {
	Path  string   `json:"path"`
	Type  *string  `json:"type,omitempty"`
	Set   []string `json:"set,omitempty"`
	Range *Range   `json:"range,omitempty"`
}

type Range struct {
	Max int `json:"max"`
	Min int `json:"min"`
}

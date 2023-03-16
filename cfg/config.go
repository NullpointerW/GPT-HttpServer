package cfg

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var Cfg = loadConfigure()

type Config struct {
	Apikey           string `json:"apikey"`
	Proxy            string `json:"proxy"`
	Port             int    `json:"port"`
	Timeout          int    `json:"timeout"`
	CharacterSetting string `json:"characterSetting"`
	TokenTTL         int    `json:"tokenTTL"`
	SecretKey        string `json:"secretKey"`
}

func loadConfigure() *Config {
	raw, err := os.ReadFile("./setting.json")
	if err != nil {
		log.Panicf("loadingConfigure failure %s:", err)
	}
	config := &Config{}
	config.Port = 8000
	config.TokenTTL = 10 * 60

	err = json.Unmarshal(raw, config)
	if err != nil {
		log.Panicf("Parsing JSON failed %s:", err)
	}
	fmt.Printf("configure:\n%+v\n", *config)
	return config
}

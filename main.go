package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Config struct {
	BackendURL string `yaml:"backend_url"`
	ListenAddr string `yaml:"listen_addr"`
	CacheSize  int    `yaml:"cache_size"`
}

func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	cache := NewCache(config.CacheSize)

	// регистрирует обработчика запросов
	http.HandleFunc("/", handleRequest(cache, config.BackendURL))

	fmt.Printf("Starting proxy on %s\n", config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}

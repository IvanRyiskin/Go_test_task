package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

// Config теги yaml указывают на соответствующие поля в yaml конфиге из-за разниыцы в регистре
type Config struct {
	BackendURL string `yaml:"backend_url"`
	ListenAddr string `yaml:"listen_addr"`
	CacheSize  int    `yaml:"cache_size"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal разбирает yaml в структуру по указанным тегам
	var config Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфиг файла: %v", err)
	}

	cache := NewCache(config.CacheSize)

	// регистрирует обработчика запросов
	// паттерн '/' обрабатывает все HTTP запросы
	http.HandleFunc("/", handleRequest(cache, config.BackendURL))

	fmt.Printf("Starting proxy on %s\n", config.ListenAddr)
	// ListenAndServe запускает HTTP сервер, который будет слушать на указанном адресе
	// nil - используется стандартный маршрутизатор ("DefaultServeMux"). Настройка в http.HandleFunc или http.Handle
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}

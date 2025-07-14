package main

import (
	"log"
	"net/http"
)

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфиг файла: %v", err)
	}

	cache := NewCache(config.Cache.Size)

	// регистрирует обработчика запросов
	// паттерн '/' обрабатывает все HTTP запросы
	// http.Handle() - для интерфейса http.Handler
	http.HandleFunc("/", handleRequest(cache, config.Backend.URL))

	log.Printf("Starting proxy on %s\n", config.Listener.Addr)
	// ListenAndServe запускает HTTP сервер, который будет слушать на указанном адресе
	// nil - используется стандартный маршрутизатор ("DefaultServeMux")
	// настройка обработчика запросов (handler) через http.HandleFunc или http.Handle
	log.Fatal(http.ListenAndServe(config.Listener.Addr, nil))
}

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

// 1. перенести урл из конфига, обрабатывать в момент запроса
// 2. добавить авторизацию на сервере (бейсик) через jrpc
// 3. ДОП. изменить хандлер на http.Handle (через интерфейс). Вынести в мейн как объект + передавать аргументом урл
// 4. ДОП. изменить авторизацию с бейсик на на токен access token (попробовать сделать со временем жизни токена refresh token)

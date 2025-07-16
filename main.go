package main

import (
	"github.com/labstack/echo/v4"
	"log"
)

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфиг файла: %v", err)
	}

	listenerAddr := config.Listener.Addr
	cacheSize := config.Cache.Size

	cache := NewCache(cacheSize)

	// Создаем сервер Echo
	e := echo.New()

	// регистрируем маршрут
	e.GET("/:domain/*", handleRequest(cache))

	// Запускаем сервер
	log.Printf("Starting proxy on %s", listenerAddr)
	log.Fatal(e.Start(listenerAddr))
}

// 1. добавить авторизацию на сервере (бейсик) через jrpc
// 2. ДОП. изменить хандлер на http.Handle (через интерфейс). Вынести в мейн как объект + передавать аргументом урл
// 3. ДОП. изменить авторизацию с бейсик на на токен access token (попробовать сделать со временем жизни токена refresh token)

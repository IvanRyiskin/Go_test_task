package main

import (
	"io"
	"net/http"
)

// обработчик http запроса. автомат реализует интерфейс Handler
// http.ResponseWriter — с помощью него мы будем отвечать клиенту.
// *http.Request — информация о пришедшем HTTP-запросе.
// Обработка тольок GET запроса. остальные обрабатываются отдельно, т.к. могут изменять состояние
func handleRequest(cache *Cache, backendURL string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		url := backendURL + request.URL.Path

		// пытаемся получить данные из кэша
		// проверка n на количество отправленных байт. Write не всегда отправляет все данные
		// Writer записывает данные как HTTP-ответ клиенту
		if cahed, ok := cache.Get(url); ok {
			n, err := writer.Write(cahed)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			if n != len(cahed) {
				http.Error(writer, "Не все данные были отправлены", http.StatusInternalServerError)
			}
			return
		}

		// если не в кеше, делаем запрос на бэк
		// Body.Close() закрывает сетевое соединение + освобождает ресурсы
		resp, err := http.Get(url)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// сохраняем в кеш
		cache.Set(url, body)

		// отправляем ответ клиенту
		n, err := writer.Write(body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		if n != len(body) {
			http.Error(writer, "Не все данные были отправлены", http.StatusInternalServerError)
		}
	}
}

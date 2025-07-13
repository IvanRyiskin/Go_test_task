package main

import (
	"io"
	"net/http"
	"strings"
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

		// формируем url для запроса + убираем пробелы
		// request.URL.Path - путь к ресурсу (endpoint)
		url := strings.TrimSpace(backendURL) + request.URL.Path

		// проверяем кеш
		// проверка n на количество отправленных байт. Write не всегда отправляет все данные
		// Writer записывает данные как HTTP-ответ клиенту, полученные от бэкенда
		if cachedItem, ok := cache.Get(url); ok {
			// отправляем заголовки
			for key, value := range cachedItem.Headers {
				// Header() возвращает type Header map[string][]string
				writer.Header()[key] = value
			}

			// отправляем cookies
			for _, cookie := range cachedItem.Cookies {
				writer.Header().Add("Set-Cookie", cookie.String())
			}

			// отправляем тело
			n, err := writer.Write(cachedItem.Body)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			if n != len(cachedItem.Body) {
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

		// для записи в кеш обязательно записать в переменную []byte
		// если кеш не нужен, можно напрямую передавать клиенту resp.Body через io.Copy (виде, стриминг)
		// Body - это io.ReadCloser, т.е. поток (stream) из которого можно читать данные по частям
		// io.ReadCloser - интерфейс, который реализует io.Reader и io.Closer
		body, err := io.ReadAll(resp.Body)
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

		// сохраняем в кеш
		cache.Set(url, body, resp.Header, resp.Cookies())

		// отправляем ответ клиенту
		for key, value := range resp.Header {
			writer.Header()[key] = value
		}

		for _, cookie := range resp.Cookies() {
			writer.Header().Add("Set-Cookie", cookie.String())
		}

		n, err := writer.Write(body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		if n != len(body) {
			http.Error(writer, "Не все данные были отправлены", http.StatusInternalServerError)
		}
	}
}

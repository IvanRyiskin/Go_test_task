package main

import (
	"io"
	"net/http"
	"strings"
)

// обработчик http запроса. автоматически реализует интерфейс Handler
// http.ResponseWriter — с помощью него мы будем отвечать клиенту.
// *http.Request — информация о пришедшем HTTP-запросе.
// http.Error() - отправляет ответ клиенту с кодом ошибки
// Обработка только GET запроса. остальные обрабатываются отдельно, т.к. могут изменять состояние
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
		if cachedItem, ok := cache.Get(url); ok {
			sendResponse(writer, cachedItem.Headers, cachedItem.Cookies, cachedItem.Body)
			return
		}

		// если не в кеше, делаем запрос на бэк
		resp, err := http.Get(url)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp != nil {
			defer func() {
				err := resp.Body.Close()
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
				}
			}()
		}

		// для записи в кеш обязательно записать в переменную []byteцвет так
		// если кеш не нужен, можно напрямую передавать клиенту resp.Body через io.Copy (виде, стриминг)
		// Body - это io.ReadCloser, т.е. поток (stream) из которого можно читать данные по частям
		// io.ReadCloser - интерфейс, который реализует io.Reader и io.Closer
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// сохраняем в кеш
		cache.Set(url, body, resp.Header, resp.Cookies())

		sendResponse(writer, resp.Header, resp.Cookies(), body)
		return
	}
}

func sendResponse(writer http.ResponseWriter, headers http.Header, cookies []*http.Cookie, body []byte) {
	// Writer записывает данные как HTTP-ответ клиенту, полученные от бэкенда
	// отправляем headers
	for key, value := range headers {
		// Header() возвращает type Header map[string][]string
		writer.Header()[key] = value
	}

	// отправляем cookies
	for _, cookie := range cookies {
		writer.Header().Add("Set-Cookie", cookie.String())
	}

	// отправляем body
	n, err := writer.Write(body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
	// проверка n на количество отправленных байт. Write не всегда отправляет все данные
	if n != len(body) {
		http.Error(writer, "Не все данные были отправлены", http.StatusInternalServerError)
	}
	return
}

package main

import (
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"strings"
)

// обработчик http запроса. автоматически реализует интерфейс Handler
// http.ResponseWriter — с помощью него мы будем отвечать клиенту.
// *http.Request — информация о пришедшем HTTP-запросе.
// http.Error() - отправляет ответ клиенту с кодом ошибки
// Обработка только GET запроса. остальные обрабатываются отдельно, т.к. могут изменять состояние
func handleRequest(cache *Cache) echo.HandlerFunc {
	return func(context echo.Context) error {
		domain := context.Param("domain")
		rest := context.Param("*")

		if !isValidDomain(domain) {
			return context.String(http.StatusBadRequest, "некорректный домен")
		}

		URL := "https://" + domain + rest

		// проверяем кеш
		if cachedItem, ok := cache.Get(URL); ok {
			sendResponse(context, cachedItem.Headers, cachedItem.Cookies, cachedItem.Body, http.StatusOK)
			return nil
		}

		// если не в кеше, делаем запрос на бэк
		resp, err := http.Get(URL)
		if err != nil {
			return context.String(http.StatusInternalServerError, err.Error())
		}

		if resp != nil {
			defer resp.Body.Close()
		}

		// для записи в кеш обязательно записать в переменную []byte
		// Body - это io.ReadCloser, т.е. поток (stream) из которого можно читать данные по частям
		// io.ReadCloser - интерфейс, который реализует io.Reader и io.Closer
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return context.String(http.StatusInternalServerError, err.Error())
		}

		// сохраняем в кеш
		cache.Set(URL, body, resp.Header, resp.Cookies(), resp.StatusCode)

		sendResponse(context, resp.Header, resp.Cookies(), body, resp.StatusCode)
		return nil
	}
}

func sendResponse(context echo.Context, headers http.Header, cookies []*http.Cookie, body []byte, statusCode int) {
	// отправляем headers
	for key, values := range headers {
		for _, v := range values {
			context.Response().Header().Add(key, v)
		}
	}

	// отправляем cookies
	for _, cookie := range cookies {
		context.Response().Header().Add("Set-Cookie", cookie.String())
	}

	// отправляем body
	context.Response().Write(body)

	// отправляем статус
	context.Response().WriteHeader(statusCode)
}

// isValidDomain проверяет, что домен выглядит корректно
func isValidDomain(domain string) bool {
	return strings.Contains(domain, ".") &&
		!strings.HasPrefix(domain, ".") &&
		!strings.HasSuffix(domain, ".") &&
		!strings.Contains(domain, "..")
}

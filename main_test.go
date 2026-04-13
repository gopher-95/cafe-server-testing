package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	for city := range cafeList {
		requests := []struct {
			count int // передаваемое значение count
			want  int // ожидаемое количество кафе в ответе
		}{
			{0, 0},
			{1, 1},
			{2, 2},
			{100, len(cafeList[city])},
		}
		for _, v := range requests {

			url := "/cafe?city=" + city + "&count=" + strconv.Itoa(v.count)

			//ответ
			response := httptest.NewRecorder()

			//запрос
			request := httptest.NewRequest("GET", url, nil)

			//регистрируем запрос и ответ
			handler.ServeHTTP(response, request)

			//проверили, что запрос прошел успешно
			require.Equal(t, http.StatusOK, response.Code)

			//получили строку, в которой через запятую указаны кафе
			body := strings.TrimSpace(response.Body.String())
			var cafes []string

			if body != "" {
				cafes = strings.Split(body, ",")
				for i := range cafes {
					cafes[i] = strings.TrimSpace(cafes[i])
				}
			}

			if v.count == 0 {
				assert.Empty(t, cafes, "пустой список")
			} else {
				assert.Equal(t, v.want, len(cafes))
			}
		}
	}
}

func TestCafeSearch(t *testing.T) {
	const city = "moscow"

	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range requests {
		url := "/cafe?city=" + city + "&search=" + v.search
		//создаем переменные для хранения запроса и ответа
		response := httptest.NewRecorder()
		request := httptest.NewRequest("GET", url, nil)
		handler.ServeHTTP(response, request)
		//проверяем работу вызовов
		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
			for i := range cafes {
				cafes[i] = strings.TrimSpace(cafes[i])
			}
		}
		//проверяем совпадает ли количество кафе
		ok := assert.Equal(t, v.wantCount, len(cafes))
		if !ok {
			http.Error(response, "не совпадает количество кафе", http.StatusBadRequest)
		}

		for _, cafe := range cafes {
			ok := strings.Contains(strings.ToLower(cafe), strings.ToLower(v.search))
			if !ok {
				http.Error(response, "произошла ошибка", http.StatusNotFound)
			}
		}

	}
}

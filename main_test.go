package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	city := "moscow"
	totalCafes := len(cafeList[city])

	requests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 0}, // Ожидаем 0 кафе при count=0
		{count: 1, want: 1}, // 1 кафе
		{count: 2, want: 2}, // 2 кафе
		{count: 100, want: min(100, totalCafes)},
	}

	for _, v := range requests {
		t.Run(fmt.Sprintf("count=%d", v.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count), nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())
			var cafes []string

			// Обрабатываем случай пустого ответа
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			// Для count=0 проверяем, что ответ пустой
			if v.count == 0 {
				assert.Empty(t, body, "Response should be empty for count=0")
				return
			}

			assert.Equal(t, v.want, len(cafes))
		})
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, v := range requests {
		t.Run(v.search, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city="+city+"&search="+v.search, nil)

			handler.ServeHTTP(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			body := strings.TrimSpace(response.Body.String())
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			assert.Equal(t, v.wantCount, len(cafes), "unexpected number of cafes")

			searchLower := strings.ToLower(v.search)
			for _, cafe := range cafes {
				cafe = strings.TrimSpace(cafe)
				assert.True(t, strings.Contains(strings.ToLower(cafe), searchLower),
					"cafe '%s' should contain '%s'", cafe, v.search)
			}
		})
	}
}

//require.Equal(t, http.StatusOK, response.Code)
//strings.TrimSpace(response.Body.String())

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

	tests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, tt := range tests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", tt.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, tt.status, response.Code)
		assert.Equal(t, tt.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, tt := range tests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", tt, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	countTests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(100, len(cafeList["moscow"]))}, // Москва имеет 5 кафе
	}

	for _, ct := range countTests {
		t.Run("count="+strconv.Itoa(ct.count), func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/cafe?city=moscow&count="+strconv.Itoa(ct.count), nil)

			mainHandle(response, req)

			require.Equal(t, http.StatusOK, response.Code)

			cafes := strings.Split(strings.TrimSpace(response.Body.String()), ",")
			if ct.count == 0 {
				assert.Len(t, cafes, 1)
				assert.Empty(t, cafes[0]) // ожидание пустой строки
			} else {
				assert.Len(t, cafes, ct.want)
			}
		})
	}
}

func TestCafeSearch(t *testing.T) {
	tests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, tt := range tests {
		t.Run(tt.search, func(t *testing.T) {
			response := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/cafe?city=moscow&search="+tt.search, nil)

			mainHandle(response, req)

			assert.Equal(t, http.StatusOK, response.Code)

			// Читаем ответ
			body := response.Body.String()

			// Разбиваем ответ на слайс строк
			result := strings.Split(strings.TrimSpace(body), ",")
			if len(body) == 0 {
				result = nil
			}

			// Проверяем количество найденных кафе
			assert.Len(t, result, tt.wantCount)

			// Проверяем, что названия кафе содержат искомую подстроку search
			for _, cafe := range result {
				if cafe != "" {
					assert.Contains(t, strings.ToLower(cafe), strings.ToLower(tt.search), "cafe %q does not contain search term %q", cafe, tt.search)
				}
			}
		})
	}
}

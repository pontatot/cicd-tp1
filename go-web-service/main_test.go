package main

import (
	"bytes"
	"encoding/json"
	"go-web-service/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCityHandler(t *testing.T) {
	// Create mock db
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	// Replace the global db with our mock
	DB = mockDB

	// Test city data
	testCity := models.City{
		ID:             1,
		DepartmentCode: "75",
		InseeCode:      "75056",
		ZipCode:        "75001",
		Name:           "Paris",
		Lat:            48.8566,
		Lon:            2.3522,
	}

	t.Run("POST city - success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO city").WithArgs(
			testCity.ID, testCity.DepartmentCode, testCity.InseeCode,
			testCity.ZipCode, testCity.Name, testCity.Lat, testCity.Lon,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		cityJSON, _ := json.Marshal(testCity)
		req := httptest.NewRequest("POST", "/city", bytes.NewBuffer(cityJSON))
		w := httptest.NewRecorder()

		CityHandler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("POST city - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/city", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		CityHandler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET cities - success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "department_code", "insee_code", "zip_code", "name", "lat", "lon"}).
			AddRow(testCity.ID, testCity.DepartmentCode, testCity.InseeCode, testCity.ZipCode, testCity.Name, testCity.Lat, testCity.Lon)

		mock.ExpectQuery("SELECT (.+) FROM city").WillReturnRows(rows)

		req := httptest.NewRequest("GET", "/city", nil)
		w := httptest.NewRecorder()

		CityHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var cities []models.City
		err := json.NewDecoder(w.Body).Decode(&cities)
		assert.NoError(t, err)
		assert.Len(t, cities, 1)
		assert.Equal(t, testCity, cities[0])
	})

	t.Run("Invalid method", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/city", nil)
		w := httptest.NewRecorder()

		CityHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHealthHandler(t *testing.T) {
	t.Run("GET health - success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/_health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("Invalid method", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/_health", nil)
		w := httptest.NewRecorder()

		HealthHandler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

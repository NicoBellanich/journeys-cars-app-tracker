package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/services"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/storage/inMemory"
)

func TestStatus(t *testing.T) {

	inMemoryTransactionFactory := inMemory.NewTransactionFactory()

	router := NewCarPool(services.NewCarPool(inMemoryTransactionFactory))

	e := NewEngineForTests(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"status":"ok"}`, w.Body.String())
}

func TestAPI(t *testing.T) {
	inMemoryTransactionFactory := inMemory.NewTransactionFactory()

	router := NewCarPool(services.NewCarPool(inMemoryTransactionFactory))

	e := NewEngineForTests(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/cars", strings.NewReader(`
	[
		{ "id": 1, "seats": 4 },
		{ "id": 2, "seats": 6 }
	]`))
	req.Header = map[string][]string{"Content-Type": {"application/json"}}
	e.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/journey", strings.NewReader(`
	{ "id": 1, "passengers": 4 }
	`))
	req.Header = map[string][]string{"Content-Type": {"application/json"}}
	e.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/locate", strings.NewReader("ID=1"))
	req.Header = map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}}
	e.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"id":1,"seats":4,"availableSeats":0}`, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/dropoff", strings.NewReader("ID=1"))
	req.Header = map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}}
	e.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func NewEngineForTests(c *CarPool) *gin.Engine {
	engine := gin.New()

	engine.GET("/status", c.GetStatus)
	engine.Any("/cars", c.PutCars)
	engine.Any("/journey", c.PostJourney)
	engine.Any("/dropoff", c.PostDropoff)
	engine.Any("/locate", c.PostLocate)

	return engine

}

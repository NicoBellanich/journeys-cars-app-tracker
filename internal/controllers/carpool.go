package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/logger"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
	model "gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/services"
)

type CarPool struct {
	service *services.CarPool
	logger  *logger.Logger
}

func NewCarPool(service *services.CarPool) *CarPool {

	c := &CarPool{
		service: service,
		logger:  logger.New("carpool-controller"),
	}

	return c
}

// GetStatus returns a basic health status for the API.
//
// GET /status
// Response: 200 OK, body {"status":"ok"}
func (c *CarPool) GetStatus(ctx *gin.Context) {
	ctx.String(http.StatusOK, `{"status":"ok"}`)
}

// PutCars resets the fleet with a new set of cars.
//
// PUT /cars
// Content-Type: application/json
// Request body: array of Car { id: number, seats: number }
// Sets availableSeats to seats for each car.
// Responses:
// - 200 OK on success
// - 400 Bad Request when duplicated id or invalid payload
// - 415/405 for wrong content type/method
func (c *CarPool) PutCars(ctx *gin.Context) {
	if ctx.Request.Method != "PUT" {
		c.logger.Error("Invalid method for cars endpoint", map[string]interface{}{
			"method":     ctx.Request.Method,
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if ctx.ContentType() != "application/json" {
		c.logger.Error("Invalid content type for cars endpoint", map[string]interface{}{
			"content_type": ctx.ContentType(),
			"request_id":   logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}

	var cars []*model.Car
	if err := ctx.BindJSON(&cars); err != nil {
		return
	}
	for _, car := range cars {
		car.AvailableSeats = car.Seats
	}

	if err := c.service.ResetCars(ctx, cars); err != nil {
		c.logger.Error("Failed to reset cars", map[string]interface{}{
			"error":      err.Error(),
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		if apiErr, ok := err.(*models.APIError); ok {
			ctx.JSON(apiErr.HTTPStatus(), apiErr)
		} else {
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}
	ctx.Status(http.StatusOK)
}

// PostJourney creates a journey request; it will be assigned to a car if
// there is capacity, otherwise it is added to a pending queue.
//
// POST /journey
// Content-Type: application/json
// Request body: Journey { id: number, passengers: number }
// Responses:
// - 200 OK on success
// - 400 Bad Request on duplicated id
// - 415/405 for wrong content type/method
func (c *CarPool) PostJourney(ctx *gin.Context) {
	if ctx.Request.Method != "POST" {
		c.logger.Error("Invalid method for journey endpoint", map[string]interface{}{
			"method":     ctx.Request.Method,
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if ctx.ContentType() != "application/json" {
		c.logger.Error("Invalid content type for journey endpoint", map[string]interface{}{
			"content_type": ctx.ContentType(),
			"request_id":   logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}

	var journey model.Journey
	if err := ctx.BindJSON(&journey); err != nil {
		return
	}
	if err := c.service.NewJourney(ctx, &journey); err != nil {
		c.logger.Error("Failed to create journey", map[string]interface{}{
			"journey_id": journey.Id,
			"error":      err.Error(),
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		if apiErr, ok := err.(*models.APIError); ok {
			ctx.JSON(apiErr.HTTPStatus(), apiErr)
		} else {
			ctx.Status(http.StatusInternalServerError)

		}
		return
	}
	ctx.Status(http.StatusOK)
}

// PostDropoff finishes a journey, freeing car seats or removing it from pending.
//
// POST /dropoff
// Content-Type: application/x-www-form-urlencoded
// Request body: ID=<journeyId>
// Responses:
// - 200 OK on success (and triggers reassignment asynchronously)
// - 204 No Content if journey had no car assigned
// - 404 Not Found if journey doesn't exist
// - 415/405 for wrong content type/method
func (c *CarPool) PostDropoff(ctx *gin.Context) {
	if ctx.Request.Method != "POST" {
		c.logger.Error("Invalid method for dropoff endpoint", map[string]interface{}{
			"method":     ctx.Request.Method,
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if ctx.ContentType() != "application/x-www-form-urlencoded" {
		c.logger.Error("Invalid content type for dropoff endpoint", map[string]interface{}{
			"content_type": ctx.ContentType(),
			"request_id":   logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}

	var dropoff struct {
		Id uint `form:"ID" binding:"required"`
	}
	if err := ctx.Bind(&dropoff); err != nil {
		return
	}

	car, err := c.service.Dropoff(ctx, dropoff.Id)
	if err != nil {
		c.logger.Error("Failed to process dropoff", map[string]interface{}{
			"journey_id": dropoff.Id,
			"error":      err.Error(),
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		if apiErr, ok := err.(*models.APIError); ok {
			ctx.JSON(apiErr.HTTPStatus(), apiErr)
		} else {
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}
	if car == nil {
		ctx.Status(http.StatusNoContent)
		return
	} else {
		if err := c.service.Reassign(ctx, car); err != nil {
			c.logger.Error("Failed to reassign car after dropoff", map[string]interface{}{
				"car_id":     car.ID,
				"journey_id": dropoff.Id,
				"error":      err.Error(),
				"request_id": logger.GetRequestID(ctx.Request.Context()),
			})
			// Don't fail the request if reassignment fails
		}
	}
	ctx.Status(http.StatusOK)
}

// PostLocate returns the car assigned to a journey, if any.
//
// POST /locate
// Content-Type: application/x-www-form-urlencoded
// Request body: ID=<journeyId>
// Responses:
// - 200 OK with car JSON when assigned
// - 204 No Content when not assigned
// - 404 Not Found if journey doesn't exist
// - 415/405 for wrong content type/method
func (c *CarPool) PostLocate(ctx *gin.Context) {
	if ctx.Request.Method != "POST" {
		c.logger.Error("Invalid method for locate endpoint", map[string]interface{}{
			"method":     ctx.Request.Method,
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusMethodNotAllowed)
		return
	}
	if ctx.ContentType() != "application/x-www-form-urlencoded" {
		c.logger.Error("Invalid content type for locate endpoint", map[string]interface{}{
			"content_type": ctx.ContentType(),
			"request_id":   logger.GetRequestID(ctx.Request.Context()),
		})
		ctx.AbortWithStatus(http.StatusUnsupportedMediaType)
		return
	}

	var locate struct {
		Id uint `form:"ID" binding:"required"`
	}
	if err := ctx.Bind(&locate); err != nil {
		return
	}

	car, err := c.service.Locate(ctx, locate.Id)
	if err != nil {
		c.logger.Error("Failed to locate journey", map[string]interface{}{
			"journey_id": locate.Id,
			"error":      err.Error(),
			"request_id": logger.GetRequestID(ctx.Request.Context()),
		})
		if apiErr, ok := err.(*models.APIError); ok {
			ctx.Status(apiErr.HTTPStatus())
		} else {
			ctx.Status(http.StatusInternalServerError)
		}
		return
	}
	if car == nil {
		ctx.Status(http.StatusNoContent)
		return
	}
	ctx.JSON(http.StatusOK, car)
}

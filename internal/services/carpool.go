package services

import (
	"context"
	"time"

	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/logger"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
)

type CarPool struct {
	transactionFactory models.TransactionFactory
	logger             *logger.Logger
}

func NewCarPool(factory models.TransactionFactory) *CarPool {
	return &CarPool{
		transactionFactory: factory,
		logger:             logger.New("carpool-service"),
	}
}

func (cp *CarPool) ResetCars(ctx context.Context, cars []*models.Car) error {
	start := time.Now()
	requestID := logger.GetRequestID(ctx)

	cp.logger.Info("Starting car reset", map[string]interface{}{
		"car_count":  len(cars),
		"request_id": requestID,
	})

	txn, err := cp.transactionFactory.Begin()
	if err != nil {
		cp.logger.Error("Failed to begin transaction for car reset", map[string]interface{}{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return models.NewAPIError(500, "Failed to begin transaction", err.Error())
	}
	defer handleTxn(txn)

	// Reset all storages
	txn.CarsStorage().ResetMemory()
	txn.JourneysStorage().ResetMemory()
	txn.PendingsStorage().ResetMemory()

	seenIDs := make(map[uint]bool)

	for _, car := range cars {
		if !car.HasValidSeats() {
			cp.logger.Error("Invalid car seats", map[string]interface{}{
				"car_id":     car.ID,
				"seats":      car.Seats,
				"request_id": requestID,
			})
			return models.ErrInvalidSeats
		}

		if seenIDs[car.ID] {
			cp.logger.Error("Duplicate car ID in request", map[string]interface{}{
				"car_id":     car.ID,
				"request_id": requestID,
			})
			return models.ErrDuplicatedID
		}
		seenIDs[car.ID] = true

		_, err := txn.CarsStorage().FindById(car.ID)
		if err != nil && err != models.ErrNotFound {
			cp.logger.Error("Error checking existing car", map[string]interface{}{
				"car_id":     car.ID,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return models.NewAPIError(500, "Failed to check existing car", err.Error())
		}

		if err := txn.CarsStorage().NewCar(car); err != nil {
			cp.logger.Error("Failed to create car", map[string]interface{}{
				"car_id":     car.ID,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return models.NewAPIError(500, "Failed to create car", err.Error())
		}
	}

	if err := txn.Commit(); err != nil {
		cp.logger.Error("Failed to commit car reset transaction", map[string]interface{}{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return models.NewAPIError(500, "Failed to commit transaction", err.Error())
	}

	cp.logger.Info("Car reset completed successfully", map[string]interface{}{
		"car_count":   len(cars),
		"duration_ms": time.Since(start).Milliseconds(),
		"request_id":  requestID,
	})

	return nil
}

func (cp *CarPool) NewJourney(ctx context.Context, journey *models.Journey) error {
	start := time.Now()
	requestID := logger.GetRequestID(ctx)

	cp.logger.Info("Starting new journey", map[string]interface{}{
		"journey_id": journey.Id,
		"passengers": journey.Passengers,
		"request_id": requestID,
	})

	txn, err := cp.transactionFactory.Begin()
	if err != nil {
		cp.logger.Error("Failed to begin transaction for new journey", map[string]interface{}{
			"journey_id": journey.Id,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return models.NewAPIError(500, "Failed to begin transaction", err.Error())
	}
	defer handleTxn(txn)

	_, err = txn.JourneysStorage().FindById(journey.Id)
	if err != nil && err != models.ErrNotFound {
		cp.logger.Error("Error checking existing journey", map[string]interface{}{
			"journey_id": journey.Id,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return models.NewAPIError(500, "Failed to check existing journey", err.Error())
	}

	seats := journey.Passengers
	cars := txn.CarsStorage().GetAllCars()
	car := findCar(cars, seats)

	if car != nil {
		car.TakeSeats(seats)
		if err := txn.CarsStorage().UpdateCar(car.ID, car); err != nil {
			cp.logger.Error("Failed to update car after assignment", map[string]interface{}{
				"car_id":     car.ID,
				"journey_id": journey.Id,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return models.NewAPIError(500, "Failed to update car", err.Error())
		}

		journey.AssignCar(car)
		if err := txn.JourneysStorage().NewJourney(journey); err != nil {
			cp.logger.Error("Failed to create journey record", map[string]interface{}{
				"journey_id": journey.Id,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return models.NewAPIError(500, "Failed to create journey", err.Error())
		}

		cp.logger.Info("Journey assigned to car", map[string]interface{}{
			"journey_id":  journey.Id,
			"car_id":      car.ID,
			"passengers":  journey.Passengers,
			"duration_ms": time.Since(start).Milliseconds(),
			"request_id":  requestID,
		})

	} else {
		if err := txn.JourneysStorage().NewJourney(journey); err != nil {
			cp.logger.Error("Failed to create pending journey record", map[string]interface{}{
				"journey_id": journey.Id,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return models.NewAPIError(500, "Failed to create journey", err.Error())
		}
		if err := txn.PendingsStorage().NewPending(journey); err != nil {
			cp.logger.Error("Failed to add journey to pending queue", map[string]interface{}{
				"journey_id": journey.Id,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return models.NewAPIError(500, "Failed to add journey to pending queue", err.Error())
		}

		cp.logger.Info("Journey added to pending queue", map[string]interface{}{
			"journey_id":  journey.Id,
			"passengers":  journey.Passengers,
			"duration_ms": time.Since(start).Milliseconds(),
			"request_id":  requestID,
		})
	}

	if err := txn.Commit(); err != nil {
		cp.logger.Error("Failed to commit journey transaction", map[string]interface{}{
			"journey_id": journey.Id,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return models.NewAPIError(500, "Failed to commit transaction", err.Error())
	}

	return nil
}

func (cp *CarPool) Dropoff(ctx context.Context, journeyId uint) (car *models.Car, err error) {

	start := time.Now()
	requestID := logger.GetRequestID(ctx)

	cp.logger.Info("Starting journey dropoff", map[string]interface{}{
		"journey_id": journeyId,
		"request_id": requestID,
	})

	txn, err := cp.transactionFactory.Begin()
	if err != nil {
		cp.logger.Error("Failed to begin transaction for dropoff", map[string]interface{}{
			"journey_id": journeyId,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return nil, models.NewAPIError(500, "Failed to begin transaction", err.Error())
	}
	defer handleTxn(txn)

	journey, err := txn.JourneysStorage().FindById(journeyId)
	if err != nil {
		cp.logger.Error("Journey not found for dropoff", map[string]interface{}{
			"journey_id": journeyId,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return nil, err
	}

	if err := txn.JourneysStorage().DeleteById(journey.Id); err != nil {
		cp.logger.Error("Failed to delete journey record", map[string]interface{}{
			"journey_id": journeyId,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return nil, models.NewAPIError(500, "Failed to delete journey", err.Error())
	}

	car = journey.AssignedTo
	if car != nil {
		car.FreeUpSeats(journey.Passengers)
		if err := txn.CarsStorage().UpdateCar(car.ID, car); err != nil {
			cp.logger.Error("Failed to update car after dropoff", map[string]interface{}{
				"car_id":     car.ID,
				"journey_id": journeyId,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return nil, models.NewAPIError(500, "Failed to update car", err.Error())
		}
		cp.logger.Info("Journey dropped off from car", map[string]interface{}{
			"journey_id":      journeyId,
			"car_id":          car.ID,
			"passengers":      journey.Passengers,
			"available_seats": car.AvailableSeats,
			"request_id":      requestID,
		})
	} else {
		if err := txn.PendingsStorage().DeleteById(journey.Id); err != nil {
			cp.logger.Error("Failed to remove journey from pending queue", map[string]interface{}{
				"journey_id": journeyId,
				"error":      err.Error(),
				"request_id": requestID,
			})
			return nil, models.NewAPIError(500, "Failed to remove journey from pending queue", err.Error())
		}
		cp.logger.Info("Pending journey dropped off", map[string]interface{}{
			"journey_id": journeyId,
			"passengers": journey.Passengers,
			"request_id": requestID,
		})
	}

	if err := txn.Commit(); err != nil {
		cp.logger.Error("Failed to commit dropoff transaction", map[string]interface{}{
			"journey_id": journeyId,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return nil, models.NewAPIError(500, "Failed to commit transaction", err.Error())
	}

	cp.logger.Info("Journey dropoff completed", map[string]interface{}{
		"journey_id":  journeyId,
		"had_car":     car != nil,
		"duration_ms": time.Since(start).Milliseconds(),
		"request_id":  requestID,
	})

	return car, nil
}

func (cp *CarPool) Reassign(ctx context.Context, car *models.Car) error {

	start := time.Now()
	requestID := logger.GetRequestID(ctx)

	cp.logger.Info("Starting car reassignment", map[string]interface{}{
		"car_id":          car.ID,
		"available_seats": car.AvailableSeats,
		"request_id":      requestID,
	})

	txn, err := cp.transactionFactory.Begin()
	if err != nil {
		cp.logger.Error("Failed to begin transaction for reassignment", map[string]interface{}{
			"car_id":     car.ID,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return models.NewAPIError(500, "Failed to begin transaction", err.Error())
	}
	defer handleTxn(txn)

	pending := txn.PendingsStorage().GetAllPendings()
	cp.logger.Debug("Retrieved pending journeys for reassignment", map[string]interface{}{
		"car_id":        car.ID,
		"pending_count": len(pending),
		"request_id":    requestID,
	})

	for _, p := range pending {
		if p.Passengers <= car.AvailableSeats {

			p.AssignedTo = car
			if err := txn.PendingsStorage().UpdatePending(p.Id, p); err != nil {
				cp.logger.Error("Failed to update pending journey", map[string]interface{}{
					"car_id":     car.ID,
					"journey_id": p.Id,
					"error":      err.Error(),
					"request_id": requestID,
				})
				return models.NewAPIError(500, "Failed to update pending journey", err.Error())
			}

			car.TakeSeats(p.Passengers)
			if err := txn.CarsStorage().UpdateCar(car.ID, car); err != nil {
				cp.logger.Error("Failed to update car after reassignment", map[string]interface{}{
					"car_id":     car.ID,
					"journey_id": p.Id,
					"error":      err.Error(),
					"request_id": requestID,
				})
				return models.NewAPIError(500, "Failed to update car", err.Error())
			}

			if err := txn.PendingsStorage().DeleteById(p.Id); err != nil {
				cp.logger.Error("Failed to remove journey from pending queue", map[string]interface{}{
					"car_id":     car.ID,
					"journey_id": p.Id,
					"error":      err.Error(),
					"request_id": requestID,
				})
				return models.NewAPIError(500, "Failed to remove journey from pending queue", err.Error())
			}

			cp.logger.Info("Journey reassigned to car", map[string]interface{}{
				"car_id":     car.ID,
				"journey_id": p.Id,
				"passengers": p.Passengers,
				"request_id": requestID,
			})
		}
	}

	if err := txn.Commit(); err != nil {
		cp.logger.Error("Failed to commit reassignment transaction", map[string]interface{}{
			"car_id":     car.ID,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return models.NewAPIError(500, "Failed to commit transaction", err.Error())
	}

	cp.logger.Info("Car reassignment completed", map[string]interface{}{
		"car_id":      car.ID,
		"duration_ms": time.Since(start).Milliseconds(),
		"request_id":  requestID,
	})

	return nil
}

func (cp *CarPool) Locate(ctx context.Context, journeyId uint) (*models.Car, error) {
	start := time.Now()
	requestID := logger.GetRequestID(ctx)

	cp.logger.Info("Starting journey location lookup", map[string]interface{}{
		"journey_id": journeyId,
		"request_id": requestID,
	})

	txn, err := cp.transactionFactory.Begin()
	if err != nil {
		cp.logger.Error("Failed to begin transaction for locate", map[string]interface{}{
			"journey_id": journeyId,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return nil, models.NewAPIError(500, "Failed to begin transaction", err.Error())
	}

	journey, err := txn.JourneysStorage().FindById(journeyId)
	if err != nil {
		cp.logger.Error("Journey not found for locate", map[string]interface{}{
			"journey_id": journeyId,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return nil, err
	}

	car := journey.AssignedTo
	if car != nil {
		cp.logger.Info("Journey located in car", map[string]interface{}{
			"journey_id":  journeyId,
			"car_id":      car.ID,
			"duration_ms": time.Since(start).Milliseconds(),
			"request_id":  requestID,
		})
	} else {
		cp.logger.Info("Journey not yet assigned to car", map[string]interface{}{
			"journey_id":  journeyId,
			"duration_ms": time.Since(start).Milliseconds(),
			"request_id":  requestID,
		})
	}

	return journey.AssignedTo, nil
}

func findCar(cars []*models.Car, seats uint) *models.Car {
	var bestCar *models.Car
	var minSeats uint = ^uint(0)

	for _, c := range cars {
		if c.AvailableSeats >= seats && c.AvailableSeats < minSeats {
			bestCar = c
			minSeats = c.AvailableSeats
		}
	}

	return bestCar
}

func handleTxn(txn models.Transaction) {
	if !txn.HasCommited() {
		txn.Rollback()
	}
}

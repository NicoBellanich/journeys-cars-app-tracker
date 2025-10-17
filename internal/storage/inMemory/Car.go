package inMemory

import (
	"sync"

	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
)

// CarStorage struct that handles inmemory car storage
// decided to be map since is faster for searching and updating than slice
type CarStorage struct {
	cars map[uint]*models.Car
	mu   sync.RWMutex
}

func NewCarStorage() *CarStorage {
	return &CarStorage{
		cars: make(map[uint]*models.Car, 0),
	}
}

func (cp *CarStorage) FindById(carId uint) (car *models.Car, err error) {
	cp.mu.RLock()
	car, exists := cp.cars[carId]
	cp.mu.RUnlock()
	if !exists {
		return nil, models.ErrNotFound
	}
	return car, nil
}

func (cp *CarStorage) GetAllCars() []*models.Car {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	var cars []*models.Car

	for _, c := range cp.cars {
		cars = append(cars, c)
	}

	return cars
}

func (cp *CarStorage) UpdateCar(carId uint, newCar *models.Car) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	car, exists := cp.cars[carId]
	if !exists {
		return models.ErrNotFound
	}

	*car = *newCar
	return nil
}

func (cp *CarStorage) NewCar(car *models.Car) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.cars[car.ID] = car
	return nil
}

func (cp *CarStorage) ResetMemory() error {
	cp.mu.Lock()
	cp.cars = make(map[uint]*models.Car, 0)
	cp.mu.Unlock()
	return nil
}

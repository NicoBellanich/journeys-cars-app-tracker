package models

type ICarStorage interface {
	NewCar(car *Car) error
	FindById(carId uint) (car *Car, err error)
	UpdateCar(carId uint, newCar *Car) error
	GetAllCars() []*Car
	ResetMemory() error
}

type IJourneyStorage interface {
	NewJourney(journey *Journey) error
	FindById(journeyId uint) (car *Journey, err error)
	DeleteById(journeyId uint) error
	UpdateJourney(journeyId uint, newJourney *Journey) error
	ResetMemory() error
}

type IPenidngStorage interface {
	NewPending(pending *Journey) error
	UpdatePending(pendingId uint, newPending *Journey) error
	DeleteById(journeyId uint) error
	GetAllPendings() []*Journey
	ResetMemory() error
}

type Transaction interface {
	CarsStorage() ICarStorage
	JourneysStorage() IJourneyStorage
	PendingsStorage() IPenidngStorage

	Commit() error
	Rollback() error

	HasCommited() bool
}

type TransactionFactory interface {
	Begin() (Transaction, error)
}

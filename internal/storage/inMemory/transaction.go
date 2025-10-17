package inMemory

import "gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"

type Transaction struct {
	carStorage     *CarStorage
	journeyStorage *JourneysStorage
	pendingStorage *PendingStorage

	carBackup     map[uint]*models.Car
	journeyBackup map[uint]*models.Journey
	pendingBackup []*models.Journey

	committed bool
}

func (u *Transaction) CarsStorage() models.ICarStorage {
	return u.carStorage
}

func (u *Transaction) JourneysStorage() models.IJourneyStorage {
	return u.journeyStorage
}

func (u *Transaction) PendingsStorage() models.IPenidngStorage {
	return u.pendingStorage
}

func (u *Transaction) Commit() error {
	u.carBackup = nil
	u.journeyBackup = nil
	u.pendingBackup = nil

	u.committed = true
	return nil
}

func (u *Transaction) HasCommited() bool {
	return u.committed
}

func (u *Transaction) Rollback() error {
	u.carStorage.cars = cloneCars(u.carBackup)
	u.journeyStorage.journeys = cloneJourneys(u.journeyBackup)
	u.pendingStorage.pending = clonePending(u.pendingBackup)
	return nil
}

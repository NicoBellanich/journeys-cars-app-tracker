package inMemory

import "gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"

type TransactionFactory struct {
	carStorage     *CarStorage
	journeyStorage *JourneysStorage
	pendingStorage *PendingStorage
}

func NewTransactionFactory() *TransactionFactory {
	return &TransactionFactory{
		carStorage:     NewCarStorage(),
		journeyStorage: NewJourneysStorage(),
		pendingStorage: NewPendingStorage(),
	}
}

func (f *TransactionFactory) Begin() (models.Transaction, error) {
	carBackup := cloneCars(f.carStorage.cars)
	journeyBackup := cloneJourneys(f.journeyStorage.journeys)
	pendingBackup := clonePending(f.pendingStorage.pending)

	return &Transaction{
		carStorage:     f.carStorage,
		journeyStorage: f.journeyStorage,
		pendingStorage: f.pendingStorage,
		carBackup:      carBackup,
		journeyBackup:  journeyBackup,
		pendingBackup:  pendingBackup,
		committed:      false,
	}, nil
}

package storage

import (
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/storage/inMemory"
)

func NewTransactionFactory(storageType string) models.TransactionFactory {
	var factory models.TransactionFactory
	switch storageType {
	case "sql":
		// db := connectToSQL() // db connection
		// factory = &SQLUnitOfWorkFactory{DB: db}
	case "memory":
		factory = inMemory.NewTransactionFactory()
	default:
		panic("unknown storage backend")
	}

	return factory
}

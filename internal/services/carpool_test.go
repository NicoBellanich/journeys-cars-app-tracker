package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	mock_models "gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/mocks"
	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
)

func TestResetCars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnFactory := mock_models.NewMockTransactionFactory(ctrl)
	txn := mock_models.NewMockTransaction(ctrl)
	carsStorage := mock_models.NewMockICarStorage(ctrl)
	journeysStorage := mock_models.NewMockIJourneyStorage(ctrl)
	pendingsStorage := mock_models.NewMockIPenidngStorage(ctrl)

	// Test valid cars are loaded and storages reset and commit is called
	cars := []*models.Car{{ID: 1, Seats: 4, AvailableSeats: 4}, {ID: 2, Seats: 6, AvailableSeats: 6}}

	txnFactory.EXPECT().Begin().Return(txn, nil)
	txn.EXPECT().CarsStorage().Return(carsStorage).AnyTimes()
	txn.EXPECT().JourneysStorage().Return(journeysStorage).AnyTimes()
	txn.EXPECT().PendingsStorage().Return(pendingsStorage).AnyTimes()

	carsStorage.EXPECT().ResetMemory().Return(nil)
	journeysStorage.EXPECT().ResetMemory().Return(nil)
	pendingsStorage.EXPECT().ResetMemory().Return(nil)

	// For each car: check not found or found then NewCar called
	carsStorage.EXPECT().FindById(uint(1)).Return(nil, models.ErrNotFound)
	carsStorage.EXPECT().NewCar(cars[0]).Return(nil)

	carsStorage.EXPECT().FindById(uint(2)).Return(nil, models.ErrNotFound)
	carsStorage.EXPECT().NewCar(cars[1]).Return(nil)

	txn.EXPECT().Commit().Return(nil)
	txn.EXPECT().HasCommited().Return(true)

	svc := NewCarPool(txnFactory)
	if err := svc.ResetCars(context.Background(), cars); err != nil {
		t.Fatalf("ResetCars returned error: %v", err)
	}
}

func TestNewJourney_AssignsToBestCar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnFactory := mock_models.NewMockTransactionFactory(ctrl)
	txn := mock_models.NewMockTransaction(ctrl)
	carsStorage := mock_models.NewMockICarStorage(ctrl)
	journeysStorage := mock_models.NewMockIJourneyStorage(ctrl)
	pendingsStorage := mock_models.NewMockIPenidngStorage(ctrl)

	journey := &models.Journey{Id: 10, Passengers: 4}

	car1 := &models.Car{ID: 1, Seats: 6, AvailableSeats: 6}
	car2 := &models.Car{ID: 2, Seats: 4, AvailableSeats: 4}

	txnFactory.EXPECT().Begin().Return(txn, nil)
	txn.EXPECT().JourneysStorage().Return(journeysStorage).AnyTimes()
	txn.EXPECT().CarsStorage().Return(carsStorage).AnyTimes()
	txn.EXPECT().PendingsStorage().Return(pendingsStorage).AnyTimes()

	journeysStorage.EXPECT().FindById(uint(10)).Return(nil, models.ErrNotFound)
	carsStorage.EXPECT().GetAllCars().Return([]*models.Car{car1, car2})
	// best fit is car2 with exactly 4 available seats
	carsStorage.EXPECT().UpdateCar(uint(2), gomock.Any()).Return(nil)
	journeysStorage.EXPECT().NewJourney(gomock.Any()).Return(nil)
	txn.EXPECT().Commit().Return(nil)
	txn.EXPECT().HasCommited().Return(true)

	svc := NewCarPool(txnFactory)
	if err := svc.NewJourney(context.Background(), journey); err != nil {
		t.Fatalf("NewJourney returned error: %v", err)
	}

	if journey.AssignedTo == nil || journey.AssignedTo.ID != 2 {
		t.Fatalf("expected journey assigned to car 2, got %+v", journey.AssignedTo)
	}
}

func TestNewJourney_PendingWhenNoCar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnFactory := mock_models.NewMockTransactionFactory(ctrl)
	txn := mock_models.NewMockTransaction(ctrl)
	carsStorage := mock_models.NewMockICarStorage(ctrl)
	journeysStorage := mock_models.NewMockIJourneyStorage(ctrl)
	pendingsStorage := mock_models.NewMockIPenidngStorage(ctrl)

	journey := &models.Journey{Id: 11, Passengers: 6}

	txnFactory.EXPECT().Begin().Return(txn, nil)
	txn.EXPECT().JourneysStorage().Return(journeysStorage).AnyTimes()
	txn.EXPECT().CarsStorage().Return(carsStorage).AnyTimes()
	txn.EXPECT().PendingsStorage().Return(pendingsStorage).AnyTimes()

	journeysStorage.EXPECT().FindById(uint(11)).Return(nil, models.ErrNotFound)
	carsStorage.EXPECT().GetAllCars().Return([]*models.Car{})
	journeysStorage.EXPECT().NewJourney(journey).Return(nil)
	pendingsStorage.EXPECT().NewPending(journey).Return(nil)
	txn.EXPECT().Commit().Return(nil)
	txn.EXPECT().HasCommited().Return(true)

	svc := NewCarPool(txnFactory)
	if err := svc.NewJourney(context.Background(), journey); err != nil {
		t.Fatalf("NewJourney returned error: %v", err)
	}

	if journey.AssignedTo != nil {
		t.Fatalf("expected journey to be pending, got assigned to car %+v", journey.AssignedTo)
	}
}

func TestDropoff_WithAssignedCar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnFactory := mock_models.NewMockTransactionFactory(ctrl)
	txn := mock_models.NewMockTransaction(ctrl)
	carsStorage := mock_models.NewMockICarStorage(ctrl)
	journeysStorage := mock_models.NewMockIJourneyStorage(ctrl)
	pendingsStorage := mock_models.NewMockIPenidngStorage(ctrl)

	car := &models.Car{ID: 1, Seats: 6, AvailableSeats: 2}
	journey := &models.Journey{Id: 20, Passengers: 4, AssignedTo: car}

	txnFactory.EXPECT().Begin().Return(txn, nil)
	txn.EXPECT().JourneysStorage().Return(journeysStorage).AnyTimes()
	txn.EXPECT().CarsStorage().Return(carsStorage).AnyTimes()
	txn.EXPECT().PendingsStorage().Return(pendingsStorage).AnyTimes()

	journeysStorage.EXPECT().FindById(uint(20)).Return(journey, nil)
	journeysStorage.EXPECT().DeleteById(uint(20)).Return(nil)
	carsStorage.EXPECT().UpdateCar(uint(1), gomock.Any()).Return(nil)
	txn.EXPECT().Commit().Return(nil)
	txn.EXPECT().HasCommited().Return(true)

	svc := NewCarPool(txnFactory)
	returnedCar, err := svc.Dropoff(context.Background(), 20)
	if err != nil {
		t.Fatalf("Dropoff returned error: %v", err)
	}
	if returnedCar == nil || returnedCar.ID != 1 {
		t.Fatalf("expected returned car 1, got %+v", returnedCar)
	}
}

func TestDropoff_PendingGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnFactory := mock_models.NewMockTransactionFactory(ctrl)
	txn := mock_models.NewMockTransaction(ctrl)
	journeysStorage := mock_models.NewMockIJourneyStorage(ctrl)
	pendingsStorage := mock_models.NewMockIPenidngStorage(ctrl)

	journey := &models.Journey{Id: 21, Passengers: 3, AssignedTo: nil}

	txnFactory.EXPECT().Begin().Return(txn, nil)
	txn.EXPECT().JourneysStorage().Return(journeysStorage).AnyTimes()
	txn.EXPECT().PendingsStorage().Return(pendingsStorage).AnyTimes()

	journeysStorage.EXPECT().FindById(uint(21)).Return(journey, nil)
	journeysStorage.EXPECT().DeleteById(uint(21)).Return(nil)
	pendingsStorage.EXPECT().DeleteById(uint(21)).Return(nil)
	txn.EXPECT().Commit().Return(nil)
	txn.EXPECT().HasCommited().Return(true)

	svc := NewCarPool(txnFactory)
	returnedCar, err := svc.Dropoff(context.Background(), 21)
	if err != nil {
		t.Fatalf("Dropoff returned error: %v", err)
	}
	if returnedCar != nil {
		t.Fatalf("expected nil car for pending dropoff, got %+v", returnedCar)
	}
}

func TestReassign_AssignsPendingThatFits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnFactory := mock_models.NewMockTransactionFactory(ctrl)
	txn := mock_models.NewMockTransaction(ctrl)
	carsStorage := mock_models.NewMockICarStorage(ctrl)
	pendingsStorage := mock_models.NewMockIPenidngStorage(ctrl)

	car := &models.Car{ID: 1, Seats: 6, AvailableSeats: 6}
	p1 := &models.Journey{Id: 30, Passengers: 2}
	p2 := &models.Journey{Id: 31, Passengers: 5}

	txnFactory.EXPECT().Begin().Return(txn, nil)
	txn.EXPECT().PendingsStorage().Return(pendingsStorage).AnyTimes()
	txn.EXPECT().CarsStorage().Return(carsStorage).AnyTimes()

	// Will attempt to assign p1 (fits), skip p2 (doesn't fit after p1)
	pendingsStorage.EXPECT().GetAllPendings().Return([]*models.Journey{p1, p2})
	pendingsStorage.EXPECT().UpdatePending(uint(30), gomock.Any()).Return(nil)
	carsStorage.EXPECT().UpdateCar(uint(1), gomock.Any()).Return(nil)
	pendingsStorage.EXPECT().DeleteById(uint(30)).Return(nil)

	txn.EXPECT().Commit().Return(nil)
	txn.EXPECT().HasCommited().Return(true)

	svc := NewCarPool(txnFactory)
	if err := svc.Reassign(context.Background(), car); err != nil {
		t.Fatalf("Reassign returned error: %v", err)
	}
}

func TestLocate_ReturnsAssignedCar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txnFactory := mock_models.NewMockTransactionFactory(ctrl)
	txn := mock_models.NewMockTransaction(ctrl)
	journeysStorage := mock_models.NewMockIJourneyStorage(ctrl)

	car := &models.Car{ID: 2, Seats: 4, AvailableSeats: 0}
	journey := &models.Journey{Id: 40, Passengers: 4, AssignedTo: car}

	txnFactory.EXPECT().Begin().Return(txn, nil)
	txn.EXPECT().JourneysStorage().Return(journeysStorage).AnyTimes()
	journeysStorage.EXPECT().FindById(uint(40)).Return(journey, nil)

	svc := NewCarPool(txnFactory)
	got, err := svc.Locate(context.Background(), 40)
	if err != nil {
		t.Fatalf("Locate returned error: %v", err)
	}
	if got == nil || got.ID != 2 {
		t.Fatalf("expected car 2, got %+v", got)
	}
}

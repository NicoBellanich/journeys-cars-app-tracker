package inMemory

import (
	"sync"

	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
)

// JourneysStorage struct that handles inmemory journey storage
// decided to be map since is faster for searching and updating than slice
type JourneysStorage struct {
	journeys map[uint]*models.Journey
	mu       sync.RWMutex
}

func NewJourneysStorage() *JourneysStorage {
	return &JourneysStorage{
		journeys: make(map[uint]*models.Journey, 0),
	}
}

func (cp *JourneysStorage) FindById(journeyId uint) (journey *models.Journey, err error) {
	cp.mu.RLock()
	journey, exists := cp.journeys[journeyId]
	cp.mu.RUnlock()
	if !exists {
		return nil, models.ErrNotFound
	}
	return journey, nil
}

func (cp *JourneysStorage) DeleteById(journeyId uint) error {
	cp.mu.Lock()
	delete(cp.journeys, journeyId)
	cp.mu.Unlock()
	return nil
}

func (cp *JourneysStorage) UpdateJourney(journeyId uint, newJourney *models.Journey) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	journey, exists := cp.journeys[journeyId]
	if !exists {
		return models.ErrNotFound
	}

	*journey = *newJourney
	return nil
}

func (cp *JourneysStorage) NewJourney(journey *models.Journey) error {
	cp.mu.Lock()
	cp.journeys[journey.Id] = journey
	cp.mu.Unlock()
	return nil
}

func (cp *JourneysStorage) ResetMemory() error {
	cp.mu.Lock()
	cp.journeys = make(map[uint]*models.Journey, 0)
	cp.mu.Unlock()
	return nil
}

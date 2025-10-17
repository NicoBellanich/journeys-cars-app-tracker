package inMemory

import (
	"sync"

	"gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"
)

// PendingStorage struct that handles inmemory pending storage
// decided to be a slice since it's important to keep the arriving order
type PendingStorage struct {
	pending []*models.Journey
	mu      sync.RWMutex
}

func NewPendingStorage() *PendingStorage {
	return &PendingStorage{
		pending: make([]*models.Journey, 0),
	}
}

func (cp *PendingStorage) GetAllPendings() []*models.Journey {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.pending
}

func (cp *PendingStorage) FindByID(pendingId uint) (journey *models.Journey, err error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	for _, pending := range cp.pending {
		if pending.Id == pendingId {
			return pending, nil
		}
	}

	return nil, models.ErrNotFound
}

func (cp *PendingStorage) UpdatePending(pendingId uint, newPending *models.Journey) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for _, pending := range cp.pending {
		if pending.Id == pendingId {
			*pending = *newPending
			return nil
		}
	}
	return models.ErrNotFound
}

func (cp *PendingStorage) DeleteById(journeyId uint) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	for i, p := range cp.pending {
		if p.Id == journeyId {
			cp.pending = append(cp.pending[:i], cp.pending[i+1:]...)
			return nil
		}
	}
	return nil
}

func (cp *PendingStorage) NewPending(pending *models.Journey) error {
	cp.mu.Lock()
	cp.pending = append(cp.pending, pending)
	cp.mu.Unlock()
	return nil
}

func (cp *PendingStorage) ResetMemory() error {
	cp.mu.Lock()
	cp.pending = make([]*models.Journey, 0)
	cp.mu.Unlock()
	return nil
}

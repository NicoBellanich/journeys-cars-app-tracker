package inMemory

import "gitlab-hiring.cabify.tech/cabify/interviewing/car-pooling-challenge-go/internal/models"

func cloneCars(src map[uint]*models.Car) map[uint]*models.Car {
	dst := make(map[uint]*models.Car)
	for k, v := range src {
		copied := *v
		dst[k] = &copied
	}
	return dst
}

func cloneJourneys(src map[uint]*models.Journey) map[uint]*models.Journey {
	dst := make(map[uint]*models.Journey)
	for k, v := range src {
		copied := *v
		dst[k] = &copied
	}
	return dst
}

func clonePending(src []*models.Journey) []*models.Journey {
	dst := []*models.Journey{}
	for _, v := range src {
		copied := *v
		dst = append(dst, &copied)
	}

	return dst
}

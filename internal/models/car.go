package models

const MAX_SEATS = 6
const MIN_SEATS = 4

type Car struct {
	ID             uint `json:"id"`
	Seats          uint `json:"seats"`
	AvailableSeats uint `json:"availableSeats"`
}

func (c *Car) HasValidSeats() bool {
	return c.Seats >= MIN_SEATS && c.Seats <= MAX_SEATS
}

func (c *Car) FreeUpSeats(amount uint) {
	c.AvailableSeats += amount
}

func (c *Car) TakeSeats(amount uint) {
	c.AvailableSeats -= amount
}

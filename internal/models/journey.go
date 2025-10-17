package models

type Journey struct {
	Id         uint `json:"id"`
	Passengers uint `json:"passengers"`
	AssignedTo *Car `json:"assignedTo"`
}

func (j *Journey) AssignCar(c *Car) {
	j.AssignedTo = c
}

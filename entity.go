package main

type cronEntity struct {
	Seconds    string `json:"seconds"`
	Minutes    string `json:"minutes"`
	Hours      string `json:"hours"`
	DayOfMonth string `json:"dayOfMonth"`
	Month      string `json:"month"`
	DayOfWeek  string `json:"dayOfWeek"`
	Year       string `json:"year"`
}


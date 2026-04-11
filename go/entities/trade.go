package data

import "time"

// Trade (also called "time and sales") represents [price, volume] pairs.
type Trade struct {
	Time   time.Time `json:"time"`  // The date and time.
	Price  float64   `json:"price"` // The price.
	Volume float64   `json:"vavue"` // The volume.
}

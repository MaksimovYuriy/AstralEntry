package entity

import "time"

type Document struct {
	ID      string
	OwnerID string
	Name    string
	Mime    string
	File    bool
	Public  bool
	Created time.Time
}

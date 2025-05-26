package model

import "time"

type InstallState struct {
	Message string    `json:"message" db:"message"`
	State   string    `json:"state" db:"state"`
	Percent int64     `json:"percent" db:"percent"`
	Time    time.Time `json:"time" db:"created_at"`
}

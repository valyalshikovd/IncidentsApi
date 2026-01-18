package domain

import "time"

type Incident struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	DangerRadiusM int       `json:"danger_radius_m"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

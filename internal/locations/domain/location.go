package domain

import "time"

type LocationCheck struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	HasDanger bool      `json:"has_danger"`
	CreatedAt time.Time `json:"created_at"`
}

type IncidentDistance struct {
	IncidentID    int64   `json:"id"`
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	DangerRadiusM int     `json:"danger_radius_m"`
	DistanceM     float64 `json:"distance_m"`
}

type CheckResult struct {
	Dangerous bool               `json:"dangerous"`
	Incidents []IncidentDistance `json:"incidents"`
}

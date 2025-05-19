package domain

import (
	"time"
)

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CategoryID  string    `json:"category_id"`
	FrameSize   string    `json:"frame_size"`
	WheelSize   string    `json:"wheel_size"`
	Color       string    `json:"color"`
	Weight      float64   `json:"weight"`
	BikeType    string    `json:"bike_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductFilter struct {
	CategoryID string
	MinPrice   *float64
	MaxPrice   *float64
	InStock    *bool
	BikeType   string
	FrameSize  string
	WheelSize  string
	Color      string
	MaxWeight  *float64
	Page       int
	PageSize   int
}

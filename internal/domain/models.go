package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Client struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Label       string     `json:"label"`
	State       string     `json:"state"`
	ClientUUID  string     `json:"client_uuid"`
	ShareLink   string     `json:"share_link,omitempty"`
	DisabledAt  *time.Time `json:"disabled_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

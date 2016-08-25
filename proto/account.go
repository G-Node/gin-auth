package proto

import "time"

type Account struct {
	URL         string       `json:"url"`
	UUID        string       `json:"uuid"`
	Login       string       `json:"login"`
	Email       *Email       `json:"email"`
	Title       *string      `json:"title"`
	FirstName   string       `json:"first_name"`
	MiddleName  *string      `json:"middle_name"`
	LastName    string       `json:"last_name"`
	Affiliation *Affiliation `json:"affiliation"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Email struct {
	Email    string `json:"email"`
	IsPublic bool   `json:"is_public"`
}

type Affiliation struct {
	Institute  string `json:"institute"`
	Department string `json:"department"`
	City       string `json:"city"`
	Country    string `json:"country"`
	IsPublic   bool   `json:"is_public"`
}

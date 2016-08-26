package proto

import "time"

// Account info json struct
type Account struct {
	URL         string       `json:"url"`
	UUID        string       `json:"uuid"`
	Login       string       `json:"login"`
	Email       *Email       `json:"email,omitempty"`
	Title       *string      `json:"title"`
	FirstName   string       `json:"first_name"`
	MiddleName  *string      `json:"middle_name"`
	LastName    string       `json:"last_name"`
	Affiliation *Affiliation `json:"affiliation,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Email json struct
type Email struct {
	Email    string `json:"email"`
	IsPublic bool   `json:"is_public"`
}

// Affiliation json struct
type Affiliation struct {
	Institute  string `json:"institute"`
	Department string `json:"department"`
	City       string `json:"city"`
	Country    string `json:"country"`
	IsPublic   bool   `json:"is_public"`
}

// SSHKey json struct
type SSHKey struct {
	URL         string    `json:"url"`
	Fingerprint string    `json:"fingerprint"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Login       string    `json:"login"`
	AccountURL  string    `json:"account_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

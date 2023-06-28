// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

type NewTodo struct {
	Text   string `json:"text"`
	UserID string `json:"userId"`
}

type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
	User *User  `json:"user"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Vehicle struct {
	ID       string    `json:"id"`
	Owner    []byte    `json:"owner"`
	Make     string    `json:"make"`
	Model    string    `json:"model"`
	Year     int       `json:"year"`
	MintTime time.Time `json:"mintTime"`
}

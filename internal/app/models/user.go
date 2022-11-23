package models

// User contains user's information
type User struct {
	Id             int
	Username       string
	HashedPassword string
}

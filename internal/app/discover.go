package app

// UserDetails contains only public details fields of a user.
type UserDetails struct {
	ID     int
	Name   string
	Gender Gender
	Age    int
}

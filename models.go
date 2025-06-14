package observer

type Role int32

const (
	RoleGuest   Role = -1
	RoleStudent Role = 0
	RoleTeacher Role = 1
	RoleAdmin   Role = 2
)

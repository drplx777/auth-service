package dtos

type UserResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Middlename string `json:"middlename,omitempty"`
	Login      string `json:"login"`
	RoleID     int    `json:"roleID"`
}

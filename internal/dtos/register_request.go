package dtos

type RegisterRequest struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Middlename string `json:"middlename,omitempty"`
	Login      string `json:"login"`
	RoleID     int    `json:"roleID"`
	Password   string `json:"password"`
}

package dtos

type LoginRequest struct {
	Login    *string `json:"login" binding:"required,min=3,max=50"`
	Password *string `json:"password" binding:"required,min=3,max=50"`
}

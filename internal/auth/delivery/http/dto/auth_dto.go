package dto

type LoginReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginRes struct {
	AccessToken string `json:"access_token"`
}

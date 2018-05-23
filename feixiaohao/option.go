package feixiaohao

type UserLoginMeta struct {
	UserID     string `json:"userid"`
	PassWD     string `json:"password"`
	IsRemember bool   `json:"isRemember"`
}

package model

type User struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	Balance   int    `json:"balance"`
	Withdrawn int    `json:"withdrawn"`
}

type RegisterUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserBalance struct {
	UserUUID  string
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

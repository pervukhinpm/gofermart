package model

type User struct {
	ID        string  `json:"id"`
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float64 `json:"balance"`
	Withdrawn float64 `json:"withdrawn"`
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
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package datatypes

// Profile struct - профиль пользователя.
type Profile struct {
	Role       string `json:"role"`
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Username   string `json:"username"`
	Schoolyear string `json:"schoolyear"`
	UID        string `json:"uid"`
}

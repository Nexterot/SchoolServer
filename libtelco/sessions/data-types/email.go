// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package dataTypes

// EmailsList struct содержит в себе список писем.
type EmailsList struct {
	Letters []Email `json:"letters"`
}

// Email struct содержит в себе одно заголовок одного письма.
type Email struct {
	Date   string `json:"date"`
	ID     int    `json:"id"`
	Author string `json:"author"`
	Title  string `json:"title"`

	// Это поле заполняется из БД.
	Unread bool `json:"unread"`
}

// EmailUser struct - пользователь электронной почты.
type EmailUser struct {
	Name string `json:"name"`

	// Это поле наверное надо заполнять из БД, но не уверен. В самом письме указываются лишь имена пользователей без ID.
	ID int `json:"id"`
}

// EmailFile struct - присоединённый к письму файл.
type EmailFile struct {
	FileName string `json:"file_name"`
	Link     string `json:"link"`

	// Эти поля нужны для скачивания файла.
	Path string
	ID   string
}

// EmailDescription struct - подробности письма.
type EmailDescription struct {
	To          []EmailUser `json:"to"`
	Copy        []EmailUser `json:"copy"`
	Description string      `json:"description"`
	Files       []EmailFile `json:"files"`
}

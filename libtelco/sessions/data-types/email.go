// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package dataTypes

// EmailsList struct содержит в себе список писем.
type EmailsList struct {
	TotalRecordCount int     `json:"TotalRecordCount"`
	ResultStatus     int     `json:"ResultStatus"`
	Result           string  `json:"Result"`
	Record           []Email `json:"Records"`
}

// Email struct содержит в себе одно заголовок одного письма.
type Email struct {
	MessageID  int    `json:"MessageId"`
	FromName   string `json:"FromName"`
	FromEOName string `json:"FromEOName"`
	Subj       string `json:"Subj"`
	Sent       string `json:"Sent"`
	Read       string `json:"Read"`
	SentTo     string `json:"SentTo"`
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

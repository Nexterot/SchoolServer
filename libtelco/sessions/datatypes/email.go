// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package datatypes

// Email adresses.

// AddressBook struct содержит в себе адресную книгу.
type AddressBook struct {
	Groups  []AddressBookGroup `json:"groups"`
	Classes []AddressBookClass `json:"classes"`
}

// AddressBookGroupUser struct -- пользователь в группе из адресной книги
type AddressBookGroupUser struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// AddressBookGroup struct -- группа пользователей в адресной книге
type AddressBookGroup struct {
	Title string                 `json:"title"`
	Users []AddressBookGroupUser `json:"users"`
}

// AddressBookClassParent struct -- родитель ученика в адресной книге
type AddressBookClassParent struct {
	Parent   string `json:"parent"`
	ParentID string `json:"parent_id"`
}

// AddressBookClassUser struct -- пользователь в классе из адресной книге
type AddressBookClassUser struct {
	Student   string                   `json:"string"`
	StudentID string                   `json:"student_id"`
	Parents   []AddressBookClassParent `json:"parents"`
}

// AddressBookClass struct -- класс (школьный) в адресной книге
type AddressBookClass struct {
	ClassName string                 `json:"class_name"`
	Users     []AddressBookClassUser `json:"users"`
	ID        int                    `json:"id"`
}

// Email.

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

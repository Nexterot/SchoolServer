// Copyright (C) 2018 Mikhail Masyagin

package dataTypes

// ForumThemesList struct содержит в себе список тем форума.
type ForumThemesList struct {
	Posts []ForumTheme `json:"posts"`
}

// ForumTheme struct содержит в себе одну тему форума.
type ForumTheme struct {
	Date       string `json:"date"`
	LastAuthor string `json:"last_author"`
	ID         int    `json:"id"`
	Creator    string `json:"creator"`
	Answers    int    `json:"answers"`
	Title      string `json:"title"`

	// Это поле заполняется из БД.
	Unread bool `json:"unread"`
}

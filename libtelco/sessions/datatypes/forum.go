// Copyright (C) 2018 Mikhail Masyagin

package datatypes

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

// ForumThemeMessages struct содержит в себе все сообщения из темы форума.
type ForumThemeMessages struct {
	Messages []ForumThemeMessage `json:"messages"`
}

// ForumThemeMessage struct содержит в себе подробности одного сообщения на форуме.
type ForumThemeMessage struct {
	Date    string `json:"date"`
	Author  string `json:"author"`
	Role    string `json:"role"`
	Message string `json:"message"`

	// Это поле заполняется из БД.
	Unread bool `json:"unread"`
}

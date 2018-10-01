package datatypes

// Posts struct - объявления.
type Posts struct {
	Posts []Post `json:"posts"`
}

// Post struct - одно объявление.
type Post struct {
	ID      int    `json:"id"`
	Unread  bool   `json:"unread"`
	Author  string `json:"author"`
	Title   string `json:"title"`
	Date    string `json:"date"`
	Message string `json:"message"`
	File    string `json:"file"`

	// Этих полей в ТЗ нет, но они нужны, чтобы обработать прикреплённый файл. Передавать их по сети вроде не нужно.
	// (Надо только записать в поле File ссылку, по которой сервер будет этот файл предоставлять)

	// Ссылка на файл
	FileLink string

	// ID файла
	FileID string

	// Имя файла
	FileName string
}

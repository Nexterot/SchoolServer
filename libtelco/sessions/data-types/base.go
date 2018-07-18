package dataTypes

/*
Список предметов.
*/

// LessonsMap struct содержит в себе список пар {предмет, id}
type LessonsMap struct {
	Data []LessonMap
}

// LessonMap struct содержит в себе имя предмета и его id.
type LessonMap struct {
	Name string
	ID   string
}

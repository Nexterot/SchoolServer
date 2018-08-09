// Copyright (C) 2018 Mikhail Masyagin & Andrey Koshelev

package dataTypes

/*
Список предметов.
*/

// LessonsMap struct содержит в себе список пар {предмет, id}
type LessonsMap struct {
	Data []LessonMap `json:"lessons"`
}

// LessonMap struct содержит в себе имя предмета и его id.
type LessonMap struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

package dataTypes

/*
Ресурсы.
*/

// Resources struct содержит в себе школьные ресурсы.
type Resources struct {
	Data []Group `json:"data"`
}

// File struct struct содержит в себе один файл в школьных ресурсах.
type File struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// Subgroup struct содержит в себе подгруппы файлов.
type Subgroup struct {
	SubgroupTitle string `json:"subgroupTitle"`
	Files         []File `json:"files"`
}

// Group struct содержит в себе одну группу файлов в школьных ресурсах.
type Group struct {
	GroupTitle string     `json:"groupTitle"`
	Files      []File     `json:"files"`
	Subgroups  []Subgroup `json:"subgroups"`
}

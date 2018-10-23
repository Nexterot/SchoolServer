// resource
package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Необходимо для gorm
)

// Resource struct представляет структуру школьного ресурса
type Resource struct {
	gorm.Model
	OwnerID   uint   // parent id
	OwnerType string // parent polymorhic type
	Name      string
	Link      string
}

// ResourceGroup struct представляет структуру группы ресурсов
type ResourceGroup struct {
	gorm.Model
	SchoolID          uint // belongs-to relation
	Title             string
	Resources         []Resource         `gorm:"polymorphic:Owner;"`
	ResourceSubgroups []ResourceSubgroup // has-many relation
}

// ResourceSubgroup struct представляет структуру подгруппы ресурсов
type ResourceSubgroup struct {
	gorm.Model
	ResourceGroupID uint
	Title           string
	Resources       []Resource `gorm:"polymorphic:Owner;"`
}

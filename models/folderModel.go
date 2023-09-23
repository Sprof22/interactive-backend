package models

import "gorm.io/gorm"

type Folder struct {
	gorm.Model
	Name  string
	Files []File `gorm:"foreignKey:FolderID"`
}

type File struct {
	gorm.Model
	Name     string
	Content  string
	FolderID uint
}

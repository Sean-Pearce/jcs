package main

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Name       string
	Password   string
	Role       string
	Preference Preference
	Files      []File
}

type File struct {
	gorm.Model
	Name         string
	Size         int64
	LastModified int64
	Sites        []Site `gorm:"many2many:file_site"`
	UserID       uint
}

type Preference struct {
	gorm.Model
	Sites []Site `gorm:"many2many:preference_site"`
}

type Site struct {
	gorm.Model
	Name string
}

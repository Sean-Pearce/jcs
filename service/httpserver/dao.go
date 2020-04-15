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
	Sites        []Site `gorm:"polymorphic:Owner;"`
	UserID       int
}

type Preference struct {
	gorm.Model
	Sites  []Site `gorm:"polymorphic:Owner;"`
	UserID int
}

type Site struct {
	gorm.Model
	Name      string
	OwnerID   int
	OwnerType string
}

package schemas

import "gorm.io/gorm"

type Paint struct {
	gorm.Model
	UserId  string
	CarName string
	FileId  string
	Bz2     bool
	Mip     bool
}

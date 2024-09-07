package dataBase

import (
	"github.com/bruxaodev/tradingPaints/internal/schemas"

	"gorm.io/gorm"
)

type DatabaseInterface interface {
	GetPaints() ([]schemas.Paint, error)
	GetPaint(userId string, carName string) (schemas.Paint, error)
	AddPaint(paint schemas.Paint) error
}

type Database struct {
	Db *gorm.DB
}

func (d *Database) GetPaints() ([]schemas.Paint, error) {
	var paints []schemas.Paint
	result := d.Db.Find(&paints)
	return paints, result.Error
}

func (d *Database) GetPaint(userId string, carName string) (schemas.Paint, error) {
	var paint schemas.Paint
	result := d.Db.Where(&schemas.Paint{
		UserId:  userId,
		CarName: carName,
	}).First(&paint)
	return paint, result.Error
}

func (d *Database) AddPaint(paint schemas.Paint) error {
	return d.Db.Create(&paint).Error
}

func (d *Database) UpdateFileId(paint schemas.Paint, file string) error {
	return d.Db.Model(&paint).Update("FileId", file).Error
}

func (d *Database) UpdateMip(paint schemas.Paint, mip bool) error {
	return d.Db.Model(&paint).Update("Mip", mip).Error
}

func (d *Database) UpdateBz2(paint schemas.Paint, bz2 bool) error {
	return d.Db.Model(&paint).Update("Bz2", bz2).Error
}

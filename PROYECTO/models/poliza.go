package models

import "time"

type Poliza struct {
	ID                uint `gorm:"primaryKey"`
	Tipo              string
	Estado            string
	FechaEmision      time.Time
	FechaFinalizacion time.Time
	UsuarioID         uint
}

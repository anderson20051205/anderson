package models

type Usuario struct {
	ID       uint `gorm:"primaryKey"`
	Nombre   string
	Correo   string `gorm:"unique"`
	Password string
	Polizas  []Poliza
}

package main

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB es la instancia global de la base de datos
var DB *gorm.DB

// Paciente representa a un paciente en el sistema
type Paciente struct {
	gorm.Model
	Nombre          string `gorm:"not null"`
	Apellido        string `gorm:"not null"`
	FechaNacimiento time.Time
	Genero          string
	Telefono        string
	Email           string
	Direccion       string
	NumeroSeguro    string
	GrupoSanguineo  string
	Alergias        string
	NotasMedicas    string
	Citas           []Cita
}

// Cita representa una cita médica
type Cita struct {
	gorm.Model
	Fecha      time.Time `gorm:"not null"`
	Motivo     string    `gorm:"not null"`
	Estado     string    `gorm:"default:'Programada'"`
	Notas      string
	PacienteID uint
}

// InicializarDB conecta con la base de datos y migra los modelos
func InicializarDB() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error al conectar a la base de datos: " + err.Error())
	}

	// Migración automática de modelos
	DB.AutoMigrate(&Paciente{}, &Cita{})
}

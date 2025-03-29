package main

import (
	"fmt"
	"log"
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
	// Información personal básica
	Nombre          string `gorm:"not null"`
	Apellido        string `gorm:"not null"`
	FechaNacimiento time.Time
	Edad            int // Campo calculado (no almacenado en BD)
	Genero          string
	EstadoCivil     string
	Ocupacion       string
	NivelEducativo  string
	Nacionalidad    string

	// Información de contacto
	Telefono                   string
	TelefonoAlternativo        string
	Email                      string
	Direccion                  string
	ContactoEmergenciaNombre   string
	ContactoEmergenciaTel      string
	ContactoEmergenciaRelacion string

	// Información administrativa
	NumeroSeguro           string
	TipoSeguro             string
	NumeroPoliza           string
	FechaVencimientoPoliza time.Time
	MedicoAsignado         string
	FechaRegistro          time.Time
	UltimaActualizacion    time.Time

	// Datos antropométricos
	Estatura              float64 // en cm
	Peso                  float64 // en kg
	IMC                   float64 // Índice de Masa Corporal
	CircunferenciaCintura float64
	CircunferenciaCadera  float64

	// Información médica general
	GrupoSanguineo                      string
	Alergias                            string
	EnfermedadesCronicas                string
	AntecedentesQuirurgicos             string
	AntecedentesFamiliares              string
	Hospitalizaciones                   string
	DiscapacidadesCondicionesEspeciales string

	// Historial ginecológico (si aplica)
	Menarca              string // Edad de primera menstruación
	FechaUltimoPeriodo   time.Time
	Embarazos            int
	Partos               int
	Cesareas             int
	Abortos              int
	MetodoAnticonceptivo string

	// Datos de vacunación
	VacunasAplicadas  string
	VacunasPendientes string
	ReaccionesVacunas string

	// Hábitos y estilo de vida
	Tabaquismo          string // No/Ex-fumador/Fumador + cantidad
	ConsumoAlcohol      string // Frecuencia y cantidad
	ActividadFisica     string // Tipo y frecuencia
	HabitosAlimenticios string
	HorasSueño          string
	ConsumoDrogas       string

	// Medicación
	MedicamentosActuales   string
	ReaccionesMedicamentos string

	// Especialidades específicas
	// Nutrición
	RestriccionesDieteticas  string
	PreferenciasAlimentarias string
	UltimoIMC                float64
	CambiosPeso              string

	// Cardiología
	PresionArterialHabitual  string
	FrecuenciaCardiacaReposo int
	Arritmias                string
	ProblemasCardiacos       string

	// Oftalmología
	ProblemasVision      string
	GraduacionLentes     string
	UltimaRevisionVisual time.Time

	// Odontología
	HistorialDental      string
	UltimaVisitaDentista time.Time
	ProblemasDentales    string

	// Psicología/Psiquiatría
	EstadoMental             string
	TratamientosPsicologicos string
	MedicacionPsiquiatrica   string

	// Notas generales
	NotasMedicas         string
	InformacionAdicional string

	// Relaciones
	Citas             []Cita
	DocumentosMedicos []DocumentoMedico `gorm:"foreignKey:PacienteID"`
	SignosVitales     []SignoVital      `gorm:"foreignKey:PacienteID"`
}

// Cita representa una cita médica
type Cita struct {
	gorm.Model
	Fecha                time.Time `gorm:"not null"`
	Motivo               string    `gorm:"not null"`
	Estado               string    `gorm:"default:'Programada'"`
	Notas                string
	Diagnostico          string
	Tratamiento          string
	MedicoAsignado       string
	Especialidad         string
	SeguimientoRequerido bool
	FechaSeguimiento     time.Time
	PacienteID           uint
}

// DocumentoMedico representa archivos médicos del paciente
type DocumentoMedico struct {
	gorm.Model
	Nombre     string
	Tipo       string // Radiografía, Análisis, Receta, etc.
	Ruta       string
	Fecha      time.Time
	Notas      string
	PacienteID uint
}

// SignoVital para guardar historial de signos vitales
type SignoVital struct {
	gorm.Model
	Fecha                  time.Time
	PresionArterial        string
	FrecuenciaCardiaca     int
	Temperatura            float64
	FrecuenciaRespiratoria int
	SaturacionOxigeno      float64
	Glucosa                float64
	Notas                  string
	PacienteID             uint
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

	// Log para verificar conexión
	log.Println("Conectado correctamente a la base de datos")

	// Migración automática de modelos
	DB.AutoMigrate(&Paciente{}, &Cita{}, &DocumentoMedico{}, &SignoVital{})
	log.Println("Migración de modelos completada")
}

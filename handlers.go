package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PageData contiene datos comunes para todas las páginas
type PageData struct {
	Title     string
	Message   string
	Error     string
	Path      string
	Paciente  Paciente
	Pacientes []Paciente
}

// IndexHandler maneja la página de inicio
func IndexHandler(c *fiber.Ctx) error {
	// Contar el total de pacientes
	var totalPacientes int64
	DB.Model(&Paciente{}).Count(&totalPacientes)

	// Contar citas para hoy
	hoy := time.Now()
	inicioDelDia := time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 0, 0, 0, 0, hoy.Location())
	finDelDia := inicioDelDia.Add(24 * time.Hour)

	var citasHoy int64
	DB.Model(&Cita{}).Where("fecha BETWEEN ? AND ?", inicioDelDia, finDelDia).Count(&citasHoy)

	return c.Render("index", fiber.Map{
		"Title":          "Sistema de Gestión de Pacientes",
		"TotalPacientes": totalPacientes,
		"CitasHoy":       citasHoy,
	})
}

// PacientesListHandler muestra la lista de pacientes
func PacientesListHandler(c *fiber.Ctx) error {
	var pacientes []Paciente
	result := DB.Find(&pacientes)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).Render("patient/patients_view", fiber.Map{
			"Title": "Lista de Pacientes",
			"Error": "Error al obtener los pacientes: " + result.Error.Error(),
		})
	}

	return c.Render("patient/patients_view", fiber.Map{
		"Title":     "Lista de Pacientes",
		"Pacientes": pacientes,
	})
}

// PacienteCreateFormHandler muestra el formulario para crear un paciente
func PacienteCreateFormHandler(c *fiber.Ctx) error {
	return c.Render("patient/patient_create", fiber.Map{
		"Title":     "Crear Nuevo Paciente",
		"Paciente":  Paciente{}, // Paciente vacío
		"ReadOnly":  false,
		"IsEditing": false,
		"Path":      c.Path(),
	})
}

// PacienteCreateHandler procesa la creación de un paciente
func PacienteCreateHandler(c *fiber.Ctx) error {
	// Crear nuevo paciente con valores básicos
	paciente := Paciente{
		Nombre:         c.FormValue("nombre"),
		Apellido:       c.FormValue("apellido"),
		Genero:         c.FormValue("genero"),
		Telefono:       c.FormValue("telefono"),
		Email:          c.FormValue("email"),
		Direccion:      c.FormValue("direccion"),
		NumeroSeguro:   c.FormValue("numeroSeguro"),
		GrupoSanguineo: c.FormValue("grupoSanguineo"),
		Alergias:       c.FormValue("alergias"),
		NotasMedicas:   c.FormValue("notasMedicas"),
	}

	fechaNacStr := c.FormValue("fechaNacimiento")
	if fechaNacStr != "" {
		fechaNac, err := time.Parse("2006-01-02", fechaNacStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).Render("patient/patient_create", fiber.Map{
				"Title":     "Crear Nuevo Paciente",
				"Error":     "Formato de fecha incorrecto",
				"Paciente":  paciente,
				"ReadOnly":  false,
				"IsEditing": false,
				"Path":      c.Path(),
			})
		}
		paciente.FechaNacimiento = fechaNac
	}

	// Guardar en la base de datos
	result := DB.Create(&paciente)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).Render("patient/patient_create", fiber.Map{
			"Title":     "Crear Nuevo Paciente",
			"Error":     "Error al crear paciente: " + result.Error.Error(),
			"Paciente":  paciente,
			"ReadOnly":  false,
			"IsEditing": false,
			"Path":      c.Path(),
		})
	}

	// Redirigir a la lista de pacientes
	return c.Redirect("/pacientes")
}

// PacienteViewHandler muestra los detalles de un paciente
func PacienteViewHandler(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "ID de paciente inválido",
		})
	}

	var paciente Paciente
	result := DB.First(&paciente, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).Render("patient/patients_view", fiber.Map{
				"Title": "Pacientes",
				"Error": "Paciente no encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "Error al obtener paciente: " + result.Error.Error(),
		})
	}

	// También obtener las citas del paciente
	var citas []Cita
	DB.Where("paciente_id = ?", id).Find(&citas)

	// Usar el mismo template que para crear/editar, pero en modo lectura
	return c.Render("patient/patient_create", fiber.Map{
		"Title":     "Detalles del Paciente",
		"Paciente":  paciente,
		"Citas":     citas,
		"ReadOnly":  true,
		"IsEditing": false,
		"Path":      c.Path(),
	})
}

// PacienteEditFormHandler muestra el formulario para editar un paciente
func PacienteEditFormHandler(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "ID de paciente inválido",
		})
	}

	var paciente Paciente
	result := DB.First(&paciente, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).Render("patient/patients_view", fiber.Map{
				"Title": "Pacientes",
				"Error": "Paciente no encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "Error al obtener paciente: " + result.Error.Error(),
		})
	}

	return c.Render("patient/patient_create", fiber.Map{
		"Title":     "Editar Paciente",
		"Paciente":  paciente,
		"ReadOnly":  false,
		"IsEditing": true,
		"Path":      c.Path(),
	})
}

// PacienteUpdateHandler procesa la actualización de un paciente
func PacienteUpdateHandler(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "ID de paciente inválido",
		})
	}

	var paciente Paciente
	result := DB.First(&paciente, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "Paciente no encontrado",
		})
	}

	// Actualizar datos del paciente
	paciente.Nombre = c.FormValue("nombre")
	paciente.Apellido = c.FormValue("apellido")
	paciente.Genero = c.FormValue("genero")
	paciente.Telefono = c.FormValue("telefono")
	paciente.Email = c.FormValue("email")
	paciente.Direccion = c.FormValue("direccion")
	paciente.NumeroSeguro = c.FormValue("numeroSeguro")
	paciente.GrupoSanguineo = c.FormValue("grupoSanguineo")
	paciente.Alergias = c.FormValue("alergias")
	paciente.NotasMedicas = c.FormValue("notasMedicas")

	// Parsear fecha de nacimiento solo si se proporciona
	fechaNacStr := c.FormValue("fechaNacimiento")
	if fechaNacStr != "" {
		fechaNac, err := time.Parse("2006-01-02", fechaNacStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).Render("patient/patient_edit", fiber.Map{
				"Title":    "Editar Paciente",
				"Paciente": paciente,
				"Error":    "Formato de fecha incorrecto",
			})
		}
		paciente.FechaNacimiento = fechaNac
	}

	// Guardar cambios
	result = DB.Save(&paciente)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).Render("patient/patient_edit", fiber.Map{
			"Title":    "Editar Paciente",
			"Paciente": paciente,
			"Error":    "Error al actualizar paciente: " + result.Error.Error(),
		})
	}

	// Redirigir a la vista de detalles
	return c.Redirect(fmt.Sprintf("/pacientes/%d", id))
}

// PacienteDeleteFormHandler muestra la confirmación para eliminar un paciente
func PacienteDeleteFormHandler(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "ID de paciente inválido",
		})
	}

	var paciente Paciente
	result := DB.First(&paciente, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "Paciente no encontrado",
		})
	}

	return c.Render("patient/patient_delete", fiber.Map{
		"Title":    "Eliminar Paciente",
		"Paciente": paciente,
	})
}

// PacienteDeleteHandler procesa la eliminación de un paciente
func PacienteDeleteHandler(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "ID de paciente inválido",
		})
	}

	// Primero eliminar las citas asociadas
	DB.Where("paciente_id = ?", id).Delete(&Cita{})

	// Luego eliminar el paciente
	result := DB.Delete(&Paciente{}, id)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).Render("patient/patients_view", fiber.Map{
			"Title": "Pacientes",
			"Error": "Error al eliminar paciente: " + result.Error.Error(),
		})
	}

	// Redirigir a la lista de pacientes
	return c.Redirect("/pacientes")
}

// PacientesFilterHandler maneja el filtrado de pacientes
func PacientesFilterHandler(c *fiber.Ctx) error {
	// Obtener los parámetros de filtrado
	var query, gender, blood, sort, direction string

	// Obtener los parámetros según el método HTTP
	if c.Method() == "POST" {
		query = c.FormValue("query")
		gender = c.FormValue("filter-gender")
		blood = c.FormValue("filter-blood")
		sort = c.FormValue("sort")
		direction = c.FormValue("direction")
	} else {
		query = c.Query("query")
		gender = c.Query("filter-gender")
		blood = c.Query("filter-blood")
		sort = c.Query("sort")
		direction = c.Query("direction")
	}

	// Limpiar espacios en blanco
	query = strings.TrimSpace(query)
	gender = strings.TrimSpace(gender)
	blood = strings.TrimSpace(blood)
	sort = strings.TrimSpace(sort)
	direction = strings.TrimSpace(direction)

	// Mostrar los parámetros recibidos en el log para depuración
	fmt.Printf("Filtro - Query: '%s', Género: '%s', Sangre: '%s', Sort: '%s', Dir: '%s'\n",
		query, gender, blood, sort, direction)

	// Crear query base
	db := DB.Model(&Paciente{})

	// Aplicar filtros de búsqueda general
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		db = db.Where(
			"LOWER(nombre) LIKE ? OR LOWER(apellido) LIKE ? OR LOWER(email) LIKE ? OR CAST(id AS TEXT) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	// Aplicar filtros específicos
	if gender != "" {
		db = db.Where("genero = ?", gender)
	}

	if blood != "" {
		db = db.Where("grupo_sanguineo = ?", blood)
	}

	// Validar columnas ordenables
	validColumns := map[string]bool{
		"id":       true,
		"nombre":   true,
		"apellido": true,
		"email":    true,
		"telefono": true,
	}

	// Aplicar ordenación
	if sort != "" && validColumns[sort] {
		orderStr := sort
		if direction == "desc" {
			orderStr += " DESC"
		} else {
			orderStr += " ASC"
		}
		db = db.Order(orderStr)
	} else {
		// Ordenación por defecto
		db = db.Order("id ASC")
		sort = "id"
		direction = "asc"
	}

	// Ejecutar consulta
	var pacientes []Paciente
	result := db.Find(&pacientes)

	fmt.Printf("Consulta completada: %d resultados encontrados\n", len(pacientes))

	if result.Error != nil {
		fmt.Printf("ERROR: %s\n", result.Error.Error())
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusInternalServerError).Render("patient/patients_filter", fiber.Map{
				"Error":     "Error al filtrar pacientes: " + result.Error.Error(),
				"Pacientes": []Paciente{},
			}, "")
		}
		return c.Status(fiber.StatusInternalServerError).Render("patient/patients_view", fiber.Map{
			"Title":     "Lista de Pacientes",
			"Error":     "Error al filtrar pacientes: " + result.Error.Error(),
			"Pacientes": []Paciente{},
		}, "")
	}

	// Para solicitudes HTMX, devolvemos solo la tabla parcial
	if c.Get("HX-Request") == "true" {
		return c.Render("patient/patients_filter", fiber.Map{
			"Pacientes":     pacientes,
			"SortColumn":    sort,
			"SortDirection": direction,
		}, "")
	}

	// Para solicitudes normales, devolvemos la página completa
	return c.Render("patient/patients_view", fiber.Map{
		"Title":     "Búsqueda de Pacientes",
		"Pacientes": pacientes,
	}, "")
}

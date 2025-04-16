package main

import (
	"fmt"
	"log"
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
	msg := c.Query("msg", "")
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).Render("patient/patients_view", fiber.Map{
			"Title":   "Lista de Pacientes",
			"Error":   "Error al obtener los pacientes: " + result.Error.Error(),
			"Message": msg,
		})
	}

	return c.Render("patient/patients_view", fiber.Map{
		"Title":     "Lista de Pacientes",
		"Pacientes": pacientes,
		"Message":   msg,
	})
}

// PacienteCreateFormHandler muestra el formulario para crear un paciente
func PacienteCreateFormHandler(c *fiber.Ctx) error {
	log.Println("[PacienteCreateFormHandler] Renderizando formulario de creación de paciente")
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
	log.Println("[PacienteCreateHandler] Procesando creación de paciente")
	// Imprimir todos los datos recibidos del formulario
	formData := make(map[string]string)
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		formData[string(key)] = string(value)
	})
	log.Printf("[DEBUG] Formulario recibido: %+v", formData)
	// Validar campos obligatorios
	nombre := strings.TrimSpace(c.FormValue("nombre"))
	apellido := strings.TrimSpace(c.FormValue("apellido"))
	genero := strings.TrimSpace(c.FormValue("genero"))
	telefono := strings.TrimSpace(c.FormValue("telefono"))
	email := strings.TrimSpace(c.FormValue("email"))
	fechaNacStr := strings.TrimSpace(c.FormValue("fechaNacimiento"))

	errores := make([]string, 0)
	if nombre == "" {
		errores = append(errores, "El nombre es obligatorio.")
	}
	if apellido == "" {
		errores = append(errores, "El apellido es obligatorio.")
	}
	if fechaNacStr == "" {
		errores = append(errores, "La fecha de nacimiento es obligatoria.")
	}
	if genero == "" {
		errores = append(errores, "El género es obligatorio.")
	}
	if telefono == "" {
		errores = append(errores, "El teléfono es obligatorio.")
	}
	if email == "" {
		errores = append(errores, "El email es obligatorio.")
	}
	if email != "" && !strings.Contains(email, "@") {
		errores = append(errores, "El email no es válido.")
	}

	// Validar duplicados por nombre/email
	var count int64
	DB.Model(&Paciente{}).Where("LOWER(nombre) = ? AND LOWER(apellido) = ?", strings.ToLower(nombre), strings.ToLower(apellido)).Count(&count)
	if count > 0 {
		errores = append(errores, "Ya existe un paciente con ese nombre y apellido.")
	}
	DB.Model(&Paciente{}).Where("LOWER(email) = ?", strings.ToLower(email)).Count(&count)
	if count > 0 {
		errores = append(errores, "Ya existe un paciente con ese email.")
	}

	// Parsear fecha de nacimiento
	var fechaNac time.Time
	if fechaNacStr != "" {
		var err error
		fechaNac, err = time.Parse("2006-01-02", fechaNacStr)
		if err != nil {
			errores = append(errores, "Formato de fecha incorrecto.")
		}
	}

	// Si hay errores, mostrar el formulario con los mensajes
	if len(errores) > 0 {
		log.Printf("[PacienteCreateHandler] Errores de validación: %v", errores)
		return c.Status(fiber.StatusBadRequest).Render("patient/patient_create", fiber.Map{
			"Title": "Crear Nuevo Paciente",
			"Error": strings.Join(errores, " "),
			"Paciente": Paciente{
				Nombre:          nombre,
				Apellido:        apellido,
				Genero:          genero,
				Telefono:        telefono,
				Email:           email,
				FechaNacimiento: fechaNac,
			},
			"ReadOnly":  false,
			"IsEditing": false,
			"Path":      c.Path(),
			"Message":   "", // Limpiar mensaje de éxito
		})
	}

	// Calcular edad
	edad := 0
	if !fechaNac.IsZero() {
		today := time.Now()
		edad = today.Year() - fechaNac.Year()
		if today.YearDay() < fechaNac.YearDay() {
			edad--
		}
	}

	// Crear nuevo paciente
	paciente := Paciente{
		Nombre:          nombre,
		Apellido:        apellido,
		Genero:          genero,
		Telefono:        telefono,
		Email:           email,
		FechaNacimiento: fechaNac,
		Edad:            edad,
		Direccion:       c.FormValue("direccion"),
		NumeroSeguro:    c.FormValue("numeroSeguro"),
		GrupoSanguineo:  c.FormValue("grupoSanguineo"),
		Alergias:        c.FormValue("alergias"),
		NotasMedicas:    c.FormValue("notasMedicas"),
	}

	log.Printf("[PacienteCreateHandler] Datos a guardar: %+v", paciente)

	result := DB.Create(&paciente)
	if result.Error != nil {
		log.Printf("[PacienteCreateHandler] Error al crear paciente: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).Render("patient/patient_create", fiber.Map{
			"Title":     "Crear Nuevo Paciente",
			"Error":     "Error al crear paciente: " + result.Error.Error(),
			"Paciente":  paciente,
			"ReadOnly":  false,
			"IsEditing": false,
			"Path":      c.Path(),
			"Message":   "", // Limpiar mensaje de éxito
		})
	}

	log.Printf("[PacienteCreateHandler] Paciente creado con ID: %d", paciente.ID)
	// Redirigir con mensaje de éxito
	return c.Redirect("/pacientes?msg=Paciente creado correctamente")
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

	// Validar campos obligatorios
	nombre := strings.TrimSpace(c.FormValue("nombre"))
	apellido := strings.TrimSpace(c.FormValue("apellido"))
	genero := strings.TrimSpace(c.FormValue("genero"))
	telefono := strings.TrimSpace(c.FormValue("telefono"))
	email := strings.TrimSpace(c.FormValue("email"))
	fechaNacStr := strings.TrimSpace(c.FormValue("fechaNacimiento"))

	errores := make([]string, 0)
	if nombre == "" {
		errores = append(errores, "El nombre es obligatorio.")
	}
	if apellido == "" {
		errores = append(errores, "El apellido es obligatorio.")
	}
	if fechaNacStr == "" {
		errores = append(errores, "La fecha de nacimiento es obligatoria.")
	}
	if genero == "" {
		errores = append(errores, "El género es obligatorio.")
	}
	if telefono == "" {
		errores = append(errores, "El teléfono es obligatorio.")
	}
	if email == "" {
		errores = append(errores, "El email es obligatorio.")
	}
	if email != "" && !strings.Contains(email, "@") {
		errores = append(errores, "El email no es válido.")
	}

	// Validar duplicados por nombre/email (excluyendo el propio paciente)
	var count int64
	DB.Model(&Paciente{}).Where("LOWER(nombre) = ? AND LOWER(apellido) = ? AND id <> ?", strings.ToLower(nombre), strings.ToLower(apellido), id).Count(&count)
	if count > 0 {
		errores = append(errores, "Ya existe un paciente con ese nombre y apellido.")
	}
	DB.Model(&Paciente{}).Where("LOWER(email) = ? AND id <> ?", strings.ToLower(email), id).Count(&count)
	if count > 0 {
		errores = append(errores, "Ya existe un paciente con ese email.")
	}

	// Parsear fecha de nacimiento
	var fechaNac time.Time
	if fechaNacStr != "" {
		var err error
		fechaNac, err = time.Parse("2006-01-02", fechaNacStr)
		if err != nil {
			errores = append(errores, "Formato de fecha incorrecto.")
		}
	}

	// Si hay errores, mostrar el formulario con los mensajes
	if len(errores) > 0 {
		return c.Status(fiber.StatusBadRequest).Render("patient/patient_create", fiber.Map{
			"Title":     "Editar Paciente",
			"Error":     strings.Join(errores, " "),
			"Paciente":  paciente,
			"ReadOnly":  false,
			"IsEditing": true,
			"Path":      c.Path(),
		})
	}

	// Calcular edad
	edad := 0
	if !fechaNac.IsZero() {
		today := time.Now()
		edad = today.Year() - fechaNac.Year()
		if today.YearDay() < fechaNac.YearDay() {
			edad--
		}
	}

	// Actualizar datos del paciente
	paciente.Nombre = nombre
	paciente.Apellido = apellido
	paciente.Genero = genero
	paciente.Telefono = telefono
	paciente.Email = email
	paciente.FechaNacimiento = fechaNac
	paciente.Edad = edad
	paciente.Direccion = c.FormValue("direccion")
	paciente.NumeroSeguro = c.FormValue("numeroSeguro")
	paciente.GrupoSanguineo = c.FormValue("grupoSanguineo")
	paciente.Alergias = c.FormValue("alergias")
	paciente.NotasMedicas = c.FormValue("notasMedicas")

	result = DB.Save(&paciente)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).Render("patient/patient_create", fiber.Map{
			"Title":     "Editar Paciente",
			"Error":     "Error al actualizar paciente: " + result.Error.Error(),
			"Paciente":  paciente,
			"ReadOnly":  false,
			"IsEditing": true,
			"Path":      c.Path(),
		})
	}

	return c.Redirect(fmt.Sprintf("/pacientes/%d?msg=Paciente actualizado correctamente", id))
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

	query = strings.TrimSpace(query)
	gender = strings.TrimSpace(gender)
	blood = strings.TrimSpace(blood)
	sort = strings.TrimSpace(sort)
	direction = strings.TrimSpace(direction)

	// Validar parámetros de ordenación
	validColumns := map[string]bool{
		"id":       true,
		"nombre":   true,
		"apellido": true,
		"email":    true,
		"telefono": true,
	}
	if sort == "" {
		sort = "id"
		direction = "asc"
	}
	if direction != "asc" && direction != "desc" {
		direction = "asc"
	}
	if !validColumns[sort] {
		// Si la columna no es válida, devolver error pero siempre con Pacientes vacío
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusBadRequest).Render("patient/patients_filter", fiber.Map{
				"Error":         "Columna de ordenación inválida.",
				"Pacientes":     []Paciente{},
				"SortColumn":    sort,
				"SortDirection": direction,
			}, "")
		}
		return c.Status(fiber.StatusBadRequest).Render("patient/patients_view", fiber.Map{
			"Title":     "Lista de Pacientes",
			"Error":     "Columna de ordenación inválida.",
			"Pacientes": []Paciente{},
		}, "")
	}

	fmt.Printf("Filtro - Query: '%s', Género: '%s', Sangre: '%s', Sort: '%s', Dir: '%s'\n",
		query, gender, blood, sort, direction)

	db := DB.Model(&Paciente{})

	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		db = db.Where(
			"LOWER(nombre) LIKE ? OR LOWER(apellido) LIKE ? OR LOWER(email) LIKE ? OR CAST(id AS TEXT) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	if gender != "" {
		db = db.Where("genero = ?", gender)
	}

	if blood != "" {
		db = db.Where("grupo_sanguineo = ?", blood)
	}

	orderStr := sort
	if direction == "desc" {
		orderStr += " DESC"
	} else {
		orderStr += " ASC"
	}
	db = db.Order(orderStr)

	var pacientes []Paciente
	result := db.Find(&pacientes)

	fmt.Printf("Consulta completada: %d resultados encontrados\n", len(pacientes))

	if result.Error != nil {
		fmt.Printf("ERROR: %s\n", result.Error.Error())
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusInternalServerError).Render("patient/patients_filter", fiber.Map{
				"Error":         "Error al filtrar pacientes: " + result.Error.Error(),
				"Pacientes":     []Paciente{},
				"SortColumn":    sort,
				"SortDirection": direction,
			}, "")
		}
		return c.Status(fiber.StatusInternalServerError).Render("patient/patients_view", fiber.Map{
			"Title":     "Lista de Pacientes",
			"Error":     "Error al filtrar pacientes: " + result.Error.Error(),
			"Pacientes": []Paciente{},
		}, "")
	}

	if c.Get("HX-Request") == "true" {
		return c.Render("patient/patients_filter", fiber.Map{
			"Pacientes":     pacientes,
			"SortColumn":    sort,
			"SortDirection": direction,
		}, "")
	}

	return c.Render("patient/patients_view", fiber.Map{
		"Title":     "Búsqueda de Pacientes",
		"Pacientes": pacientes,
	}, "")
}

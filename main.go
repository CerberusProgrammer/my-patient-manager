package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No se pudo cargar el archivo .env")
	}

	// Inicializar la base de datos
	InicializarDB()

	// Configurar el motor de plantillas
	engine := html.New("./templates", ".html")
	engine.Reload(true) // Habilitar recarga en desarrollo
	engine.Debug(true)  // Habilitar modo depuración

	// Añadir funciones personalizadas a las plantillas
	engine.AddFunc("hasPrefix", func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	})

	// Crear la aplicación Fiber
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layout/layout",
	})

	// Middleware para añadir la ruta actual a los datos de la plantilla
	app.Use(func(c *fiber.Ctx) error {
		// Añadir la ruta actual a los datos de la plantilla
		c.Locals("Path", c.Path())
		return c.Next()
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Static("/static", "./static")

	// Rutas
	app.Get("/", IndexHandler)

	// Rutas de Pacientes
	app.Get("/pacientes", PacientesListHandler)
	app.Get("/pacientes/crear", PacienteCreateFormHandler)
	app.Post("/pacientes/crear", PacienteCreateHandler)
	app.Get("/pacientes/:id", PacienteViewHandler)
	app.Get("/pacientes/:id/editar", PacienteEditFormHandler)
	app.Post("/pacientes/:id/editar", PacienteUpdateHandler)
	app.Get("/pacientes/:id/eliminar", PacienteDeleteFormHandler)
	app.Post("/pacientes/:id/eliminar", PacienteDeleteHandler)
	app.Get("/pacientes/filtrar", PacientesFilterHandler)
	app.Post("/pacientes/filtrar", PacientesFilterHandler)

	// Iniciar el servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "4332"
	}
	log.Fatal(app.Listen(":" + port))
}

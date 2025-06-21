package main

import (
	"PROYECTO/db"
	"PROYECTO/handlers"
	"PROYECTO/models"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	// Conectar y migrar base de datos
	db.Connect()
	db.DB.AutoMigrate(&models.Usuario{}, &models.Poliza{})

	r := mux.NewRouter()

	// Ruta principal, puede mostrar home o index
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index.html", nil)
	}).Methods("GET")

	// Crear usuario y poliza
	r.HandleFunc("/crear", handlers.CrearUsuarioYPoliza).Methods("POST")

	// Agregar póliza a usuario existente (esto es lo que te faltaba)
	r.HandleFunc("/agregar-poliza", handlers.AgregarPoliza).Methods("POST")

	// Login: formulario y procesamiento
	r.HandleFunc("/login", handlers.LoginForm).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Renovar póliza (POST)
	r.HandleFunc("/renovar-poliza", handlers.RenovarPoliza).Methods("POST")

	// Cancelar póliza (POST)
	r.HandleFunc("/cancelar-poliza", handlers.CancelarPoliza).Methods("POST")

	// Imprimir PDF (GET)
	r.HandleFunc("/imprimir-pdf", handlers.ImprimirPolizasPDF).Methods("GET")

	// Servir archivos estáticos (CSS, JS, imágenes)
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Println("Servidor corriendo en http://localhost:3000")
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}

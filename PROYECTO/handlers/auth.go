package handlers

import (
	"PROYECTO/db"
	"PROYECTO/models"
	"html/template"
	"net/http"
)

// Aquí cargo todas las plantillas HTML desde la carpeta templates
var tmpl = template.Must(template.ParseGlob("templates/*.html"))

// Esta función muestra el formulario de inicio de sesión al usuario
func LoginForm(w http.ResponseWriter, r *http.Request) {
	// Renderizo el archivo login.html sin pasarle ningún dato dinámico
	tmpl.ExecuteTemplate(w, "login.html", nil)
}

// Esta función procesa el inicio de sesión después de que el usuario envía el formulario
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Primero leo los datos que el usuario envió en el formulario
	r.ParseForm()

	// Creo una variable para guardar el usuario que voy a buscar en la base de datos
	var user models.Usuario

	// Consulto en la base de datos si existe un usuario con ese correo y contraseña
	db.DB.Where("correo = ? AND password = ?", r.FormValue("correo"), r.FormValue("password")).First(&user)

	// Si el usuario no existe (ID = 0), muestro un error de credenciales inválidas
	if user.ID == 0 {
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Si el usuario es válido, busco todas las pólizas asociadas a ese usuario
	var polizas []models.Poliza
	db.DB.Where("usuario_id = ?", user.ID).Find(&polizas)

	// Creo una estructura que contiene tanto al usuario como a sus pólizas para enviarla al HTML
	data := struct {
		Usuario models.Usuario
		Polizas []models.Poliza
	}{Usuario: user, Polizas: polizas}

	// Finalmente muestro la plantilla polizas.html con los datos cargados del usuario y sus pólizas
	tmpl.ExecuteTemplate(w, "polizas.html", data)
}

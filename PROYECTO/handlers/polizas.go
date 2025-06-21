package handlers

import (
	"PROYECTO/db"
	"PROYECTO/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// En esta estructura defino cómo envío respuestas JSON al frontend
type JsonResponse struct {
	Mensaje  string `json:"mensaje"`
	Error    string `json:"error,omitempty"`
	PolizaID uint   `json:"poliza_id,omitempty"`
}

// En esta función me encargo de crear un usuario nuevo y al mismo tiempo generarle una póliza automáticamente.
func CrearUsuarioYPoliza(w http.ResponseWriter, r *http.Request) {
	// Primero obtengo los datos que el usuario ingresó en el formulario
	r.ParseForm()
	usuario := models.Usuario{
		Nombre:   r.FormValue("nombre"),
		Correo:   r.FormValue("correo"),
		Password: r.FormValue("password"),
	}
	// Guardo el nuevo usuario en la base de datos
	db.DB.Create(&usuario)

	// Defino la fecha de emisión como la fecha actual
	fechaEmision := time.Now()
	// La fecha de finalización la pongo un año después de la emisión
	fechaFinalizacion := fechaEmision.AddDate(1, 0, 0)

	// Creo la póliza asociándola al usuario recién creado
	poliza := models.Poliza{
		Tipo:              r.FormValue("tipo"),
		Estado:            r.FormValue("estado"),
		FechaEmision:      fechaEmision,
		FechaFinalizacion: fechaFinalizacion,
		UsuarioID:         usuario.ID,
	}
	// Guardo la póliza en la base de datos
	db.DB.Create(&poliza)

	// Redirijo al usuario a la página principal
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Esta función me sirve para validar que una fecha tenga el formato correcto y que sea una fecha válida.
func validarFechaFormato(valor string) (time.Time, error) {
	// Intento convertir la cadena de texto a una fecha
	fecha, err := time.Parse("2006-01-02", valor)
	if err != nil {
		return time.Time{}, fmt.Errorf("formato incorrecto, use YYYY-MM-DD")
	}

	// Divido la fecha en año, mes y día para validar cada parte
	parts := strings.Split(valor, "-")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("formato incorrecto, use YYYY-MM-DD")
	}

	// Convierto cada parte en número
	anio, _ := strconv.Atoi(parts[0])
	mes, _ := strconv.Atoi(parts[1])
	dia, _ := strconv.Atoi(parts[2])

	// Valido que el año esté dentro del rango permitido
	if anio < 2000 || anio > 2050 {
		return time.Time{}, fmt.Errorf("el año debe estar entre 2000 y 2050")
	}
	// Valido que el mes sea válido
	if mes < 1 || mes > 12 {
		return time.Time{}, fmt.Errorf("el mes debe estar entre 1 y 12")
	}
	// Valido que el día sea válido
	if dia < 1 || dia > 31 {
		return time.Time{}, fmt.Errorf("el día debe estar entre 1 y 31")
	}

	// Si todo está correcto, devuelvo la fecha
	return fecha, nil
}

// Esta función la uso para agregar una nueva póliza a un usuario que ya existe en el sistema
func AgregarPoliza(w http.ResponseWriter, r *http.Request) {
	// Primero obtengo todos los datos que vienen del formulario
	r.ParseForm()

	// Aquí convierto el ID del usuario a entero para poder trabajarlo
	usuarioID, err := strconv.Atoi(r.FormValue("usuario_id"))
	// Verifico si el ID del usuario es válido (que sea un número mayor a 0)
	if err != nil || usuarioID <= 0 {
		// Si no es válido, envío un error en formato JSON
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JsonResponse{Error: "ID de usuario inválido"})
		return
	}

	// Ahora valido la fecha de emisión usando la función que yo mismo hice antes, pero esta vez con el campo "fecha_inicio"
	fechaEmision, err := validarFechaFormato(r.FormValue("fecha_emision"))
	if err != nil {
		// Si la fecha no tiene el formato correcto o es inválida, muestro un error en JSON
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JsonResponse{Error: "Fecha de inicio inválida: " + err.Error()})
		return
	}

	// También valido la fecha de finalización de la póliza usando el campo "fecha_final"
	fechaFinalizacion, err := validarFechaFormato(r.FormValue("fecha_finalizacion"))
	if err != nil {
		// Si está mal, informo el error en JSON
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JsonResponse{Error: "Fecha de finalización inválida: " + err.Error()})
		return
	}

	// Me aseguro de que la fecha de finalización sea después de la fecha de emisión
	if fechaFinalizacion.Before(fechaEmision) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(JsonResponse{Error: "La fecha de finalización debe ser posterior a la de inicio"})
		return
	}

	// Una vez que todo es válido, creo la póliza con los datos recibidos
	poliza := models.Poliza{
		Tipo:              r.FormValue("tipo"),
		Estado:            r.FormValue("estado"),
		FechaEmision:      fechaEmision,
		FechaFinalizacion: fechaFinalizacion,
		UsuarioID:         uint(usuarioID),
	}

	// Guardo la póliza en la base de datos
	db.DB.Create(&poliza)

	// Configuro la respuesta para que sea en formato JSON (esto facilita integrarlo con el frontend)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Devuelvo un mensaje de éxito en formato JSON junto con el ID de la nueva póliza
	json.NewEncoder(w).Encode(JsonResponse{
		Mensaje:  "Póliza agregada correctamente",
		PolizaID: poliza.ID,
	})
}

// Esta función la uso para renovar una póliza, pero solo si ya está vencida.
func RenovarPoliza(w http.ResponseWriter, r *http.Request) {
	// Obtengo los datos del formulario
	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))

	// Busco la póliza en la base de datos
	var poliza models.Poliza
	result := db.DB.First(&poliza, id)
	// Si no existe, muestro un error
	if result.Error != nil {
		http.Error(w, "Póliza no encontrada", http.StatusNotFound)
		return
	}

	// Verifico que la póliza realmente haya vencido
	if time.Now().Before(poliza.FechaFinalizacion) {
		http.Error(w, "No se puede renovar aún. La póliza sigue vigente.", http.StatusForbidden)
		return
	}

	// Actualizo las fechas para renovar por un año más
	poliza.FechaEmision = time.Now()
	poliza.FechaFinalizacion = poliza.FechaEmision.AddDate(1, 0, 0)
	poliza.Estado = "activa"

	// Guardo los cambios en la base de datos
	db.DB.Save(&poliza)
	// Redirijo a la página principal
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Esta función la implementé para cancelar una póliza, pero solo si ya venció.
func CancelarPoliza(w http.ResponseWriter, r *http.Request) {
	// Obtengo los datos del formulario
	r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("id"))

	// Busco la póliza correspondiente
	var poliza models.Poliza
	result := db.DB.First(&poliza, id)
	// Si no existe, devuelvo un error
	if result.Error != nil {
		http.Error(w, "Póliza no encontrada", http.StatusNotFound)
		return
	}

	// Me aseguro de que ya haya vencido antes de cancelar
	if time.Now().Before(poliza.FechaFinalizacion) {
		http.Error(w, "No se puede cancelar aún. La póliza sigue vigente.", http.StatusForbidden)
		return
	}

	// Cambio el estado de la póliza a "cancelada"
	poliza.Estado = "cancelada"
	db.DB.Save(&poliza)
	// Redirijo al usuario a la página principal
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// aquí presento un servicio de descarga de reportes en formato PDF
// aquí se imprime el PDF de las pólizas de cada usuario tomando en cuenta esto http://localhost:3000/imprimir-pdf?usuario=1  el número del usuario varía dependiendo del ID de cada usuario
func ImprimirPolizasPDF(w http.ResponseWriter, r *http.Request) {
	// Obtengo el ID de usuario desde la URL
	usuarioID, err := strconv.Atoi(r.URL.Query().Get("usuario"))
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
		return
	}

	// Busco al usuario junto con sus pólizas usando preload
	var usuario models.Usuario
	err = db.DB.Preload("Polizas").First(&usuario, usuarioID).Error
	if err != nil {
		http.Error(w, "Usuario no encontrado", http.StatusNotFound)
		return
	}

	// Creo un nuevo PDF en formato A4, orientación vertical
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Polizas de %s", usuario.Nombre))
	pdf.Ln(12)

	// Configuro encabezados de tabla con tamaños para cada columna
	pdf.SetFont("Arial", "B", 12)
	colWidths := []float64{50, 30, 50, 50}
	headers := []string{"Tipo", "Estado", "Fecha Emision", "Fecha Finalizacion"}

	// Imprimo encabezados
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 10, header, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Imprimo los datos de cada póliza
	pdf.SetFont("Arial", "", 12)
	for _, poliza := range usuario.Polizas {
		pdf.CellFormat(colWidths[0], 10, poliza.Tipo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[1], 10, poliza.Estado, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[2], 10, poliza.FechaEmision.Format("02-01-2006"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[3], 10, poliza.FechaFinalizacion.Format("02-01-2006"), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// Configuro encabezados para que el navegador lo entienda como PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=polizas.pdf")

	// Envío el PDF generado al cliente
	err = pdf.Output(w)
	if err != nil {
		http.Error(w, "Error al generar PDF", http.StatusInternalServerError)
	}
}

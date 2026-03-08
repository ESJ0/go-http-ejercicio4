package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ─── Modelos ──────────────────────────────────────────────────────────────────

type PowerStats struct {
	Intelligence int `json:"intelligence"`
	Strength     int `json:"strength"`
	Speed        int `json:"speed"`
	Durability   int `json:"durability"`
	Power        int `json:"power"`
	Combat       int `json:"combat"`
}

type Appearance struct {
	Gender    string  `json:"gender"`
	Race      string  `json:"race"`
	HeightCm  float64 `json:"height_cm"`
	WeightKg  float64 `json:"weight_kg"`
	EyeColor  string  `json:"eye_color"`
	HairColor string  `json:"hair_color"`
}

type Biography struct {
	FullName        string `json:"full_name"`
	AlterEgos       string `json:"alter_egos"`
	PlaceOfBirth    string `json:"place_of_birth"`
	FirstAppearance string `json:"first_appearance"`
}

type Character struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	RealName   string     `json:"real_name"`
	Alignment  string     `json:"alignment"`
	Publisher  string     `json:"publisher"`
	PowerStats PowerStats `json:"powerstats"`
	Appearance Appearance `json:"appearance"`
	Biography  Biography  `json:"biography"`
	Team       string     `json:"team"`
}

// ─── Respuestas estándar ──────────────────────────────────────────────────────

type ErrorResponse struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Mensaje string `json:"mensaje"`
}

type SuccessResponse struct {
	Mensaje string      `json:"mensaje"`
	Datos   interface{} `json:"datos,omitempty"`
}

// ─── Constantes ───────────────────────────────────────────────────────────────

const archivoData = "./data/characters.json"
const puerto = ":24585"

var personajes []Character

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	cargarDatos()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/characters", charactersHandler)
	http.HandleFunc("/api/characters/", characterByIDHandler)

	log.Printf("DC Characters API iniciada en el puerto %s", puerto)
	log.Fatal(http.ListenAndServe(puerto, nil))
}

// ─── Carga y persistencia ─────────────────────────────────────────────────────

func cargarDatos() {
	archivo, err := os.ReadFile(archivoData)
	if err != nil {
		log.Fatal("Error al leer el archivo de datos:", err)
	}
	if err = json.Unmarshal(archivo, &personajes); err != nil {
		log.Fatal("Error al parsear el JSON:", err)
	}
	log.Printf("Se cargaron %d personajes correctamente", len(personajes))
}

func guardarDatos() error {
	datos, err := json.MarshalIndent(personajes, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar: %w", err)
	}
	return os.WriteFile(archivoData, datos, 0644)
}

func generarSiguienteID() int {
	maxID := 0
	for _, p := range personajes {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	return maxID + 1
}

func buscarPorID(id int) (Character, int, bool) {
	for i, p := range personajes {
		if p.ID == id {
			return p, i, true
		}
	}
	return Character{}, -1, false
}

// ─── Helpers JSON ─────────────────────────────────────────────────────────────

func escribirJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println("Error al codificar la respuesta JSON:", err)
	}
}

func escribirError(w http.ResponseWriter, status int, errMsg, detalle string) {
	escribirJSON(w, status, ErrorResponse{
		Status:  status,
		Error:   errMsg,
		Mensaje: detalle,
	})
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

// GET /api/ping
func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		escribirError(w, http.StatusMethodNotAllowed, "Method Not Allowed",
			"Este endpoint solo acepta el método GET")
		return
	}
	escribirJSON(w, http.StatusOK, SuccessResponse{Mensaje: "pong"})
}

// /api/characters — colección (GET, POST)
func charactersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetAll(w, r)
	case http.MethodPost:
		handleCreate(w, r)
	default:
		escribirError(w, http.StatusMethodNotAllowed, "Method Not Allowed",
			fmt.Sprintf("El método %s no está permitido. Métodos soportados: GET, POST", r.Method))
	}
}

// /api/characters/{id} — recurso individual (GET, PUT, PATCH, DELETE)
func characterByIDHandler(w http.ResponseWriter, r *http.Request) {
	segmento := strings.TrimPrefix(r.URL.Path, "/api/characters/")
	segmento = strings.Trim(segmento, "/")

	if segmento == "" {
		charactersHandler(w, r)
		return
	}

	id, err := strconv.Atoi(segmento)
	if err != nil {
		escribirError(w, http.StatusBadRequest, "Bad Request",
			fmt.Sprintf("'%s' no es un ID válido. Debe ser un número entero.", segmento))
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetByID(w, id)
	case http.MethodPut:
		handleReplace(w, r, id)
	case http.MethodPatch:
		handleUpdate(w, r, id)
	case http.MethodDelete:
		handleDelete(w, id)
	default:
		escribirError(w, http.StatusMethodNotAllowed, "Method Not Allowed",
			fmt.Sprintf("El método %s no está permitido. Métodos soportados: GET, PUT, PATCH, DELETE", r.Method))
	}
}

// ─── GET /api/characters  (con filtros combinados) ────────────────────────────

func handleGetAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Filtro por id (query param)
	if idParam := q.Get("id"); idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			escribirError(w, http.StatusBadRequest, "Bad Request",
				"El parámetro 'id' debe ser un número entero")
			return
		}
		p, _, encontrado := buscarPorID(id)
		if !encontrado {
			escribirError(w, http.StatusNotFound, "Not Found",
				fmt.Sprintf("No se encontró ningún personaje con id=%d", id))
			return
		}
		escribirJSON(w, http.StatusOK, p)
		return
	}

	resultado := make([]Character, len(personajes))
	copy(resultado, personajes)

	// Filtro: name (contiene, case-insensitive)
	if v := q.Get("name"); v != "" {
		filtrado := []Character{}
		buscar := strings.ToLower(v)
		for _, p := range resultado {
			if strings.Contains(strings.ToLower(p.Name), buscar) {
				filtrado = append(filtrado, p)
			}
		}
		resultado = filtrado
	}

	// Filtro: alignment (good / bad / neutral)
	if v := q.Get("alignment"); v != "" {
		filtrado := []Character{}
		for _, p := range resultado {
			if strings.EqualFold(p.Alignment, v) {
				filtrado = append(filtrado, p)
			}
		}
		resultado = filtrado
	}

	// Filtro: team
	if v := q.Get("team"); v != "" {
		filtrado := []Character{}
		buscar := strings.ToLower(v)
		for _, p := range resultado {
			if strings.Contains(strings.ToLower(p.Team), buscar) {
				filtrado = append(filtrado, p)
			}
		}
		resultado = filtrado
	}

	// Filtro: gender
	if v := q.Get("gender"); v != "" {
		filtrado := []Character{}
		for _, p := range resultado {
			if strings.EqualFold(p.Appearance.Gender, v) {
				filtrado = append(filtrado, p)
			}
		}
		resultado = filtrado
	}

	// Filtro: min_intelligence
	if v := q.Get("min_intelligence"); v != "" {
		val, err := strconv.Atoi(v)
		if err != nil {
			escribirError(w, http.StatusBadRequest, "Bad Request",
				"El parámetro 'min_intelligence' debe ser un número entero")
			return
		}
		filtrado := []Character{}
		for _, p := range resultado {
			if p.PowerStats.Intelligence >= val {
				filtrado = append(filtrado, p)
			}
		}
		resultado = filtrado
	}

	// Filtro: min_combat
	if v := q.Get("min_combat"); v != "" {
		val, err := strconv.Atoi(v)
		if err != nil {
			escribirError(w, http.StatusBadRequest, "Bad Request",
				"El parámetro 'min_combat' debe ser un número entero")
			return
		}
		filtrado := []Character{}
		for _, p := range resultado {
			if p.PowerStats.Combat >= val {
				filtrado = append(filtrado, p)
			}
		}
		resultado = filtrado
	}

	if resultado == nil {
		resultado = []Character{}
	}

	escribirJSON(w, http.StatusOK, resultado)
}

// ─── GET /api/characters/{id} ────────────────────────────────────────────────

func handleGetByID(w http.ResponseWriter, id int) {
	p, _, encontrado := buscarPorID(id)
	if !encontrado {
		escribirError(w, http.StatusNotFound, "Not Found",
			fmt.Sprintf("No se encontró ningún personaje con id=%d", id))
		return
	}
	escribirJSON(w, http.StatusOK, p)
}

// ─── POST /api/characters ────────────────────────────────────────────────────

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var nuevo Character

	if err := json.NewDecoder(r.Body).Decode(&nuevo); err != nil {
		escribirError(w, http.StatusBadRequest, "Bad Request",
			"El cuerpo de la solicitud debe ser un JSON válido")
		return
	}

	if errores := validarPersonaje(nuevo); len(errores) > 0 {
		escribirJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"status":  422,
			"error":   "Validation Failed",
			"mensaje": "Errores de validación: " + strings.Join(errores, "; "),
		})
		return
	}

	nuevo.ID = generarSiguienteID()
	if nuevo.Publisher == "" {
		nuevo.Publisher = "DC Comics"
	}
	if nuevo.Biography.FirstAppearance == "" {
		nuevo.Biography.FirstAppearance = "Primera aparición: " + time.Now().Format("2006-01-02")
	}

	personajes = append(personajes, nuevo)

	if err := guardarDatos(); err != nil {
		escribirError(w, http.StatusInternalServerError, "Internal Server Error",
			"El personaje fue creado en memoria pero no pudo persistirse: "+err.Error())
		return
	}

	escribirJSON(w, http.StatusCreated, nuevo)
}

// ─── PUT /api/characters/{id} ────────────────────────────────────────────────

func handleReplace(w http.ResponseWriter, r *http.Request, id int) {
	_, idx, encontrado := buscarPorID(id)
	if !encontrado {
		escribirError(w, http.StatusNotFound, "Not Found",
			fmt.Sprintf("No se encontró ningún personaje con id=%d", id))
		return
	}

	var reemplazo Character
	if err := json.NewDecoder(r.Body).Decode(&reemplazo); err != nil {
		escribirError(w, http.StatusBadRequest, "Bad Request",
			"El cuerpo de la solicitud debe ser un JSON válido")
		return
	}

	if errores := validarPersonaje(reemplazo); len(errores) > 0 {
		escribirJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"status":  422,
			"error":   "Validation Failed",
			"mensaje": "Errores de validación: " + strings.Join(errores, "; "),
		})
		return
	}

	reemplazo.ID = id
	personajes[idx] = reemplazo

	if err := guardarDatos(); err != nil {
		escribirError(w, http.StatusInternalServerError, "Internal Server Error",
			"El personaje fue actualizado en memoria pero no pudo persistirse: "+err.Error())
		return
	}

	escribirJSON(w, http.StatusOK, reemplazo)
}

// ─── PATCH /api/characters/{id} ──────────────────────────────────────────────

func handleUpdate(w http.ResponseWriter, r *http.Request, id int) {
	existente, idx, encontrado := buscarPorID(id)
	if !encontrado {
		escribirError(w, http.StatusNotFound, "Not Found",
			fmt.Sprintf("No se encontró ningún personaje con id=%d", id))
		return
	}

	var parche map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&parche); err != nil {
		escribirError(w, http.StatusBadRequest, "Bad Request",
			"El cuerpo de la solicitud debe ser un JSON válido")
		return
	}

	if len(parche) == 0 {
		escribirError(w, http.StatusBadRequest, "Bad Request",
			"El cuerpo del PATCH no puede estar vacío")
		return
	}

	// Campos de nivel raíz
	if v, ok := parche["name"].(string); ok && v != "" {
		existente.Name = v
	}
	if v, ok := parche["real_name"].(string); ok && v != "" {
		existente.RealName = v
	}
	if v, ok := parche["alignment"].(string); ok {
		existente.Alignment = v
	}
	if v, ok := parche["publisher"].(string); ok && v != "" {
		existente.Publisher = v
	}
	if v, ok := parche["team"].(string); ok {
		existente.Team = v
	}

	// PowerStats anidado
	if ps, ok := parche["powerstats"].(map[string]interface{}); ok {
		if v, err := toInt(ps["intelligence"]); err == nil {
			existente.PowerStats.Intelligence = v
		}
		if v, err := toInt(ps["strength"]); err == nil {
			existente.PowerStats.Strength = v
		}
		if v, err := toInt(ps["speed"]); err == nil {
			existente.PowerStats.Speed = v
		}
		if v, err := toInt(ps["durability"]); err == nil {
			existente.PowerStats.Durability = v
		}
		if v, err := toInt(ps["power"]); err == nil {
			existente.PowerStats.Power = v
		}
		if v, err := toInt(ps["combat"]); err == nil {
			existente.PowerStats.Combat = v
		}
	}

	// Appearance anidado
	if ap, ok := parche["appearance"].(map[string]interface{}); ok {
		if v, ok2 := ap["gender"].(string); ok2 {
			existente.Appearance.Gender = v
		}
		if v, ok2 := ap["race"].(string); ok2 {
			existente.Appearance.Race = v
		}
		if v, err := toFloat(ap["height_cm"]); err == nil {
			existente.Appearance.HeightCm = v
		}
		if v, err := toFloat(ap["weight_kg"]); err == nil {
			existente.Appearance.WeightKg = v
		}
		if v, ok2 := ap["eye_color"].(string); ok2 {
			existente.Appearance.EyeColor = v
		}
		if v, ok2 := ap["hair_color"].(string); ok2 {
			existente.Appearance.HairColor = v
		}
	}

	personajes[idx] = existente

	if err := guardarDatos(); err != nil {
		escribirError(w, http.StatusInternalServerError, "Internal Server Error",
			"El personaje fue modificado en memoria pero no pudo persistirse: "+err.Error())
		return
	}

	escribirJSON(w, http.StatusOK, existente)
}

// ─── DELETE /api/characters/{id} ─────────────────────────────────────────────

func handleDelete(w http.ResponseWriter, id int) {
	_, idx, encontrado := buscarPorID(id)
	if !encontrado {
		escribirError(w, http.StatusNotFound, "Not Found",
			fmt.Sprintf("No se encontró ningún personaje con id=%d", id))
		return
	}

	eliminado := personajes[idx]
	personajes = append(personajes[:idx], personajes[idx+1:]...)

	if err := guardarDatos(); err != nil {
		escribirError(w, http.StatusInternalServerError, "Internal Server Error",
			"El personaje fue eliminado en memoria pero no pudo persistirse: "+err.Error())
		return
	}

	escribirJSON(w, http.StatusOK, SuccessResponse{
		Mensaje: fmt.Sprintf("El personaje '%s' (id=%d) fue eliminado correctamente", eliminado.Name, id),
		Datos:   eliminado,
	})
}

// ─── Validación ───────────────────────────────────────────────────────────────

func validarPersonaje(p Character) []string {
	var errores []string

	if strings.TrimSpace(p.Name) == "" {
		errores = append(errores, "el campo 'name' es obligatorio")
	}
	if strings.TrimSpace(p.RealName) == "" {
		errores = append(errores, "el campo 'real_name' es obligatorio")
	}
	if p.Alignment != "" && p.Alignment != "good" && p.Alignment != "bad" && p.Alignment != "neutral" {
		errores = append(errores, "el campo 'alignment' debe ser 'good', 'bad' o 'neutral'")
	}

	// PowerStats rango 0–100
	stats := map[string]int{
		"intelligence": p.PowerStats.Intelligence,
		"strength":     p.PowerStats.Strength,
		"speed":        p.PowerStats.Speed,
		"durability":   p.PowerStats.Durability,
		"power":        p.PowerStats.Power,
		"combat":       p.PowerStats.Combat,
	}
	for campo, valor := range stats {
		if valor < 0 || valor > 100 {
			errores = append(errores, fmt.Sprintf("powerstats.%s debe estar entre 0 y 100", campo))
		}
	}

	if p.Appearance.HeightCm < 0 {
		errores = append(errores, "appearance.height_cm no puede ser negativo")
	}
	if p.Appearance.WeightKg < 0 {
		errores = append(errores, "appearance.weight_kg no puede ser negativo")
	}

	return errores
}

// ─── Conversores de tipo ──────────────────────────────────────────────────────

func toInt(v interface{}) (int, error) {
	if v == nil {
		return 0, fmt.Errorf("valor nulo")
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case string:
		return strconv.Atoi(n)
	}
	return 0, fmt.Errorf("no se puede convertir %T a int", v)
}

func toFloat(v interface{}) (float64, error) {
	if v == nil {
		return 0, fmt.Errorf("valor nulo")
	}
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case string:
		return strconv.ParseFloat(n, 64)
	}
	return 0, fmt.Errorf("no se puede convertir %T a float64", v)
}
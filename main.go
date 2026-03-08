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

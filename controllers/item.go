package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"crud/models"
)

type PartidoController struct {
	PartidoModel *models.PartidoModel
}

// INSERTAR
func (c *PartidoController) Create(w http.ResponseWriter, r *http.Request) {

	var partido models.Partido

	err := json.NewDecoder(r.Body).Decode(&partido)

	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	err = c.PartidoModel.Create(&partido)

	if err != nil {
		http.Error(w, "Error al crear el partido", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(partido)
}

// VER TODOS
func (c *PartidoController) GetAll(w http.ResponseWriter, r *http.Request) {

	partidos, err := c.PartidoModel.GetAll()

	if err != nil {
		http.Error(w, "Error al obtener los partidos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(partidos)
}

// VER UNO
func (c *PartidoController) GetByID(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	partido, err := c.PartidoModel.GetByID(id)

	if err == sql.ErrNoRows {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Error al obtener el partido", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(partido)
}

// ACTUALIZAR
func (c *PartidoController) Update(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	var partido models.Partido

	err := json.NewDecoder(r.Body).Decode(&partido)

	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	updatedPartido, err := c.PartidoModel.Update(id, &partido)

	if err == sql.ErrNoRows {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Error al actualizar el partido", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(updatedPartido)
}

// BORRAR
func (c *PartidoController) Delete(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	err := c.PartidoModel.Delete(id)

	if err == sql.ErrNoRows {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Error al eliminar el partido", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		map[string]string{
			"message": "Partido eliminado correctamente",
		},
	)
}

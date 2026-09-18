package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
)

type PredictionController struct {
	WorkerName string
	ScriptPath string
}

type PredictionRecord struct {
	ID                int64   `json:"id"`
	HomeAvgGoalsLast5 float64 `json:"home_avg_goals_last5"`
}

type PredictionBatchRequest struct {
	Records []PredictionRecord `json:"records"`
}

type Prediction struct {
	ID         int64   `json:"id"`
	Prediction float64 `json:"prediction"`
}

type PythonBatchResponse struct {
	Predictions []Prediction `json:"predictions"`
}

type WorkerBatchResponse struct {
	Worker      string       `json:"worker"`
	Processed   int          `json:"processed"`
	Predictions []Prediction `json:"predictions"`
}

func (c *PredictionController) PredictBatch(
	w http.ResponseWriter,
	r *http.Request,
) {

	var request PredictionBatchRequest

	err := json.NewDecoder(
		r.Body,
	).Decode(&request)

	if err != nil {

		http.Error(
			w,
			"JSON inválido",
			http.StatusBadRequest,
		)

		return
	}

	if len(request.Records) == 0 {

		http.Error(
			w,
			"No se recibieron registros",
			http.StatusBadRequest,
		)

		return
	}

	requestJSON, err := json.Marshal(
		request,
	)

	if err != nil {

		http.Error(
			w,
			"Error preparando datos",
			http.StatusInternalServerError,
		)

		return
	}

	cmd := exec.Command(
		"python3",
		c.ScriptPath,
	)

	cmd.Stdin = bytes.NewReader(
		requestJSON,
	)

	output, err := cmd.Output()

	if err != nil {

		log.Printf(
			"Error ejecutando Python: %v",
			err,
		)

		http.Error(
			w,
			"Error ejecutando modelo",
			http.StatusInternalServerError,
		)

		return
	}

	var pythonResponse PythonBatchResponse

	err = json.Unmarshal(
		output,
		&pythonResponse,
	)

	if err != nil {

		log.Printf(
			"Error leyendo respuesta Python: %v",
			err,
		)

		http.Error(
			w,
			"Error procesando predicción",
			http.StatusInternalServerError,
		)

		return
	}

	response := WorkerBatchResponse{
		Worker:      c.WorkerName,
		Processed:   len(request.Records),
		Predictions: pythonResponse.Predictions,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		response,
	)
}

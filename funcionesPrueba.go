func heartbeatHandler(w http.REsponseWriter, r *htpp.REquest) {
	w.WriteHeader(statusCode: http.statusOK)
	w.Write([]byte("gateway alive"))
}

func main() {
	routes []Route, err := loadROutes(path: "routes.json")
	if err !=nil
}

// Revisar que bakend está libre para mandarle cosas

func checkHeartbeat() { // Hacer una petición para verificar
	// REvisar que estén vivos

	// Target
}

func statHeartbeatMonitor() {
	// Cada cuanto revisar que están vivos
	// Paralelizar cada cuanto están vivos
}

// Tengo que avisar que está vivo

// 1 go rutina que es el main que invoca a otra go rutina 



// backend guardar "algo"

// Necesitamos tener un cache

func (s *Store) Create(name string) Item {

}

// El middleware nunca debe de interactuar con la base de datos
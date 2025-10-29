package routes

import (
	"net/http"
	"pitch/controllers"
)

func Web() {
	http.HandleFunc("/", controllers.Pitch)
}

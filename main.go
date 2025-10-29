package main

import (
	"net/http"
	"pitch/routes"
)

func main() {

	routes.Web()

	http.ListenAndServe(":8082", nil)
}

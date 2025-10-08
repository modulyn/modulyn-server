package controllers

import (
	"modulyn/pkg/db"
	"modulyn/pkg/server"
	"net/http"
)

type Controller interface {
	EventsController(w http.ResponseWriter, r *http.Request)
	FeaturesController(w http.ResponseWriter, r *http.Request)
	FeatureByIdController(w http.ResponseWriter, r *http.Request)
	ProjectsController(w http.ResponseWriter, r *http.Request)
	ProjectByIdControllers(w http.ResponseWriter, r *http.Request)
	EnvironmentsController(w http.ResponseWriter, r *http.Request)
	EnvironmentByIdControllers(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	conn  db.Conn
	store server.Store
}

func New(conn db.Conn, store server.Store) Controller {
	return &controller{
		conn:  conn,
		store: store,
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
}

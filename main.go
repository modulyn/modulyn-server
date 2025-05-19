package main

import (
	"log"
	"modulyn/pkg/controllers"
	"modulyn/pkg/db"
	"modulyn/pkg/server"
	"net/http"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	store := server.NewStore()

	conn, err := db.InitDB()
	if err != nil {
		log.Fatalln("Failed to initialize database: ", err)
	}
	defer conn.Close()

	controllers := controllers.New(conn, store)

	// events
	http.HandleFunc("/events", controllers.EventsController)

	// features
	http.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}/features", controllers.FeaturesController)

	http.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}/features/{featureId}", controllers.FeatureByIdControllers)

	// projects
	http.HandleFunc("/api/v1/projects", controllers.ProjectsController)

	http.HandleFunc("/api/v1/projects/{projectId}", controllers.ProjectByIdControllers)

	// environments
	http.HandleFunc("/api/v1/projects/{projectId}/environments", controllers.EnvironmentsController)

	http.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}", controllers.EnvironmentByIdControllers)

	http.ListenAndServe(":8080", nil)
}

package main

import (
	"log"
	"modulyn/pkg/controllers"
	"modulyn/pkg/db"
	"modulyn/pkg/middlewares"
	"modulyn/pkg/server"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	store := server.NewStore()

	conn, err := db.InitDB(os.Getenv("ENABLE_SQL_LOGGING") == "true")
	if err != nil {
		log.Fatalln("Failed to initialize database: ", err)
	}
	defer conn.Close()

	controllers := controllers.New(conn, store)

	mux := http.NewServeMux()

	// events
	mux.HandleFunc("/events", controllers.EventsController)

	// features
	mux.HandleFunc("/api/v1/projects/{projectId}/features", controllers.FeaturesController)

	mux.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}/features", controllers.FeaturesByEnvironmentIDController)

	mux.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}/features/{featureId}", controllers.FeatureByIdControllers)

	// projects
	mux.HandleFunc("/api/v1/projects", controllers.ProjectsController)

	mux.HandleFunc("/api/v1/projects/{projectId}", controllers.ProjectByIdControllers)

	// environments
	mux.HandleFunc("/api/v1/projects/{projectId}/environments", controllers.EnvironmentsController)

	mux.HandleFunc("/api/v1/projects/{projectId}/environments/{environmentId}", controllers.EnvironmentByIdControllers)

	handler := middlewares.CorrelationMiddleware(mux)

	http.ListenAndServe(":8080", handler)
}

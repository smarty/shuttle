package main

import (
	"log"
	"net/http"

	"github.com/smarty/shuttle/v2"
	"github.com/smarty/shuttle/v2/sample/app"
	"github.com/smarty/shuttle/v2/sample/inputs"
)

func main() {
	// You can use any routing mechanism you'd like. For this sample, we'll be using net/http.ServeMux.
	router := http.NewServeMux()

	// About as simple as a route definition gets:
	router.Handle("/add", shuttle.NewHandler(
		shuttle.Options.InputModel(func() shuttle.InputModel { return inputs.NewAddition() }),
		shuttle.Options.Processor(func() shuttle.Processor { return app.NewProcessor() }),
	))

	// This route expects JSON in the request body:
	router.Handle("/sub", shuttle.NewHandler(
		shuttle.Options.InputModel(func() shuttle.InputModel { return inputs.NewSubtraction() }),
		shuttle.Options.Processor(func() shuttle.Processor { return app.NewProcessor() }),
		shuttle.Options.DeserializeJSON(true),
	))

	// Nothing interesting to see here...
	address := "localhost:8080"
	log.Printf("Listening on %s", address)
	err := http.ListenAndServe(address, router)
	if err != nil {
		log.Fatal(err)
	}
}

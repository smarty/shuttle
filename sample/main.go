package main

import (
	"log"
	"net/http"

	"github.com/smarty/shuttle"
	"github.com/smarty/shuttle/sample/app"
	"github.com/smarty/shuttle/sample/inputs"
)

func main() {
	router := http.NewServeMux()
	router.Handle("/add", shuttle.NewHandler(
		shuttle.Options.InputModel(func() shuttle.InputModel { return inputs.NewAddition() }),
		shuttle.Options.Processor(func() shuttle.Processor { return app.NewProcessor() }),
	))
	router.Handle("/sub", shuttle.NewHandler(
		shuttle.Options.InputModel(func() shuttle.InputModel { return inputs.NewSubtraction() }),
		shuttle.Options.Processor(func() shuttle.Processor { return app.NewProcessor() }),
		shuttle.Options.DeserializeJSON(true),
	))
	address := "localhost:8080"
	log.Printf("Listening on %s", address)
	err := http.ListenAndServe(address, router)
	if err != nil {
		log.Fatal(err)
	}
}

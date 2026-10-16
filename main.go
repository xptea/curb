// Curb's executable only reads build metadata and starts the application.
package main

import (
	_ "embed"
	"encoding/json"
	"log"

	"curb/internal/app"
)

//go:embed mygo.json
var appConfig []byte

//go:embed resources/icon.png
var appIcon []byte

func main() {
	var config struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(appConfig, &config); err != nil {
		log.Fatal(err)
	}
	if err := app.Run(config.Version, appIcon); err != nil {
		log.Fatal(err)
	}
}

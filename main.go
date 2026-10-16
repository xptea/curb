// Curb's executable only reads build metadata and starts the application.
package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"os"

	"curb/internal/app"
	"curb/internal/update"
)

//go:embed mygo.json
var appConfig []byte

//go:embed resources/icon.png
var appIcon []byte

func main() {
	if len(os.Args) == 3 && os.Args[1] == update.HelperFlag {
		if err := update.RunHelper(os.Args[2]); err != nil {
			log.Fatal(err)
		}
		return
	}
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

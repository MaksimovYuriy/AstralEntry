package main

import (
	"log"

	"github.com/maksimovyuriy/astralentry/internal/app"
)

func main() {
	if err := app.Migrate(); err != nil {
		log.Fatal(err)
	}
}

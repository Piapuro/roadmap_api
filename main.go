package main

import (
	"log"

	"github.com/your-name/roadmap/api/dicontainer"
)

func main() {
	container, err := dicontainer.New()
	if err != nil {
		log.Fatalf("failed to initialize container: %v", err)
	}

	if err := container.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

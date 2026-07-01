package main

import (
	"fmt"
	"log"

	"github.com/kostapo/tasktracker/task-tracker-backend/internal/config"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg.Database.DSN())

}

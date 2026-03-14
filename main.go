package main

import (
	"mlnck/api"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env" + ".local")
	_ = godotenv.Load()

	r := api.SetupRouter()
	_ = r.Run()
}

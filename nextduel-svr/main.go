package main

import (
	"log"
	"nextduel-svr/routes"
	"nextduel-svr/server"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	s := server.NewServer()
	routes.RegisterRoutes(s)
	log.Fatal(s.Start(":" + port))
}

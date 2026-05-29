package main

import (
	"log"
	"nextduel-svr/gamepkg"
	"nextduel-svr/routes"
	"nextduel-svr/server"
)

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	s := server.NewServer()
	g := gamepkg.NewGameManager(db)

	s.SetDatabase(db)
	s.SetGameManager(g)
	routes.RegisterRoutes(s)

	log.Fatal(s.Start())
}

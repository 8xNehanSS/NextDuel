package health

import "nextduel-svr/gamepkg"

type HealthServices struct {
	Game *gamepkg.GameManager
}

func New(game *gamepkg.GameManager) *HealthServices {
	return &HealthServices{
		Game: game,
	}
}

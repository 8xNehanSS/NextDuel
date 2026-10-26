package health

import "nextduel-svr/gamepkg"

type HealthServices struct {
	Game    *gamepkg.GameManager
	Version string
}

func New(game *gamepkg.GameManager, version string) *HealthServices {
	return &HealthServices{
		Game:    game,
		Version: version,
	}
}

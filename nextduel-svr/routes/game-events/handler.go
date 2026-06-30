package game_events

import "nextduel-svr/gamepkg"

type GameEventServices struct {
	Game *gamepkg.GameManager
}

func New(game *gamepkg.GameManager) *GameEventServices {
	return &GameEventServices{
		Game: game,
	}
}

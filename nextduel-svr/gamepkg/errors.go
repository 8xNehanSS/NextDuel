package gamepkg

import "errors"

// GameError is a client-facing error with a stable machine-readable code.
type GameError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *GameError) Error() string { return e.Message }

func newError(code, msg string) *GameError { return &GameError{Code: code, Message: msg} }

var (
	ErrAlreadyQueued   = newError("already_queued", "You are already in the matchmaking queue.")
	ErrNotQueued       = newError("not_queued", "You are not in the matchmaking queue.")
	ErrAlreadyInRoom   = newError("already_in_room", "You are already in a room.")
	ErrNotInRoom       = newError("not_in_room", "You are not in a room.")
	ErrRoomNotFound    = newError("room_not_found", "That room does not exist.")
	ErrRoomFull        = newError("room_full", "That room is already full.")
	ErrRoomClosed      = newError("room_closed", "That room has closed.")
	ErrMatchStarted    = newError("match_started", "The match has already started.")
	ErrMatchNotActive  = newError("match_not_active", "There is no question to answer right now.")
	ErrWrongRound      = newError("wrong_round", "That answer is for a different round.")
	ErrInvalidAnswer   = newError("invalid_answer", "That is not one of the choices.")
	ErrAlreadyAnswered = newError("already_answered", "You have already answered this round.")
	ErrBadRequest      = newError("bad_request", "Malformed request.")
	ErrUnknownAction   = newError("unknown_action", "Unknown action.")
	ErrUnknownMode     = newError("unknown_mode", "That game is not available.")
	ErrShuttingDown    = newError("shutting_down", "The server is restarting. Please try again shortly.")
)

func toErrorPayload(err error) *GameError {
	var ge *GameError
	if errors.As(err, &ge) {
		return ge
	}
	return newError("internal", "Something went wrong.")
}

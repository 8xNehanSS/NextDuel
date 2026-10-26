package gamepkg

import (
	"nextduel-svr/questions"
	"nextduel-svr/utils"
	"time"
)

// Recorder persists match history. Room calls it while holding its lock, so
// implementations must not call back into the room. Errors are the
// recorder's to log; gameplay never stops because persistence failed.
//
// The production implementation is repository.MatchRecorder.
type Recorder interface {
	// StartMatch is called once both players are ready and questions are
	// loaded. playerIDs are in slot order.
	StartMatch(roomID string, bestOf int, playerIDs []string) (matchID uint)
	StartRound(matchID uint, roundNo int, q questions.Question) (roundID uint)
	RecordAnswer(matchID, roundID uint, playerID, answer string, at time.Time)
	ResolveRound(matchID, roundID uint, roundNo int, winnerID *string, results []PlayerRoundResult)
	// CompleteMatch finalises the match and updates lifetime player stats.
	CompleteMatch(matchID uint, status utils.MatchStatus, winnerID *string, playerIDs []string, reason string)
	// AbortMatch marks a match aborted: no result, no stats, hidden from
	// history (e.g. server shutdown).
	AbortMatch(matchID uint, reason string)
}

type noopRecorder struct{}

func (noopRecorder) StartMatch(string, int, []string) uint                            { return 0 }
func (noopRecorder) StartRound(uint, int, questions.Question) uint                    { return 0 }
func (noopRecorder) RecordAnswer(uint, uint, string, string, time.Time)               {}
func (noopRecorder) ResolveRound(uint, uint, int, *string, []PlayerRoundResult)       {}
func (noopRecorder) CompleteMatch(uint, utils.MatchStatus, *string, []string, string) {}
func (noopRecorder) AbortMatch(uint, string)                                          {}

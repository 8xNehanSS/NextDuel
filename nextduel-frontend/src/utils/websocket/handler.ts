import { GameAction, type ServerEventType, type ServerEvents } from "./events";
import { useGameStore } from "../../store/gameStore";
import { useUserStore } from "../../store/userStore";
import { navigationService } from "../navigationService";
import { appEventBus } from "../eventbus";
import type { RoomCloseReason, RoomState } from "../../types/protocol";

type Handlers = { [K in ServerEventType]?: (payload: ServerEvents[K]) => void };

const game = () => useGameStore.getState();
const me = () => useUserStore.getState().username;

const closeReasons: Record<RoomCloseReason, string> = {
    opponent_left: "Your opponent left before the duel started.",
    questions_unavailable: "Questions could not be loaded. Please try again.",
    server_shutdown: "The server is restarting. Please rejoin in a moment.",
};

function enterRoom(room: RoomState) {
    game().enteredRoom(room);
    navigationService.go("/game");
}

export const handlers: Handlers = {
    [GameAction.ERROR]: (err) => {
        game().answerRejected();
        appEventBus.fire("notification", { message: err?.message ?? "Something went wrong.", type: "error" });
    },

    [GameAction.WAITING_QUEUE]: ({ mode, position }) => game().queued(mode, position),
    [GameAction.QUEUE_LEFT]: () => game().reset(),

    [GameAction.MATCH_FOUND]: (room) => {
        appEventBus.fire("notification", { message: "Opponent found!", type: "success" });
        enterRoom(room);
    },
    [GameAction.ROOM_CREATED]: enterRoom,
    [GameAction.ROOM_JOINED]: enterRoom,
    [GameAction.ROOM_STATE]: (room) => game().roomUpdated(room),

    [GameAction.PLAYER_LEFT]: ({ playerId }) => {
        if (playerId !== me()) {
            appEventBus.fire("notification", { message: `${playerId} left the room.`, type: "info" });
        }
    },

    [GameAction.ROOM_DESTROYED]: ({ reason }) => {
        appEventBus.fire("notification", { message: closeReasons[reason] ?? "The room was closed.", type: "info" });
        game().reset();
        navigationService.go("/lobby");
    },

    [GameAction.DUEL_START]: ({ maxRounds }) => game().starting(maxRounds),
    [GameAction.ROUND_START]: (q) => game().roundStarted(q),
    [GameAction.TIMER_UPDATE]: ({ remaining }) => game().tick(remaining),
    [GameAction.INTERMISSION_TIMER_UPDATE]: ({ remaining }) => game().tick(remaining),
    [GameAction.DUEL_ANSWER]: ({ answer }) => game().answerAccepted(answer),

    [GameAction.DUEL_FAST]: ({ playerId, remaining }) => {
        if (playerId !== me()) {
            game().opponentAnsweredFast(remaining);
        }
    },

    [GameAction.ROUND_RESULT]: (result) => game().roundResolved(result),
    [GameAction.INTERMISSION]: ({ nextRound, remaining }) => game().intermission(nextRound, remaining),
    [GameAction.GAME_OVER]: (result) => game().finished(result),
};

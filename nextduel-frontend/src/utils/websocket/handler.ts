import { GameAction } from "./events";
import { useGameStore } from "../../store/gameStore";
import type { GameStart } from "../../types/game";
import { navigationService } from "../navigationService";
import { appEventBus } from "../eventbus";

type Handler = (payload: any) => void;

export const handlers: Partial<Record<GameAction, Handler>> = {
  [GameAction.ERROR]: (payload) => {
    console.error("Error from server:", payload);
  },
  [GameAction.WAITING_QUEUE]: () => {
    useGameStore.getState().setPhase("waiting");
  },
  [GameAction.MATCH_FOUND]: () => {
    useGameStore.getState().setPhase("matched");
  },  
  [GameAction.ROOM_CREATED]: (payload) => {
    useGameStore.getState().setConnected(true);
    useGameStore.getState().setRoom(payload);
    navigationService.go("/game");
  },
  [GameAction.ROOM_JOINED]: (payload) => {
    useGameStore.getState().setRoom(payload);
  },
  [GameAction.PLAYER_JOINED]: (payload) => {
    console.log("Player joined:", payload);
  },
  [GameAction.PLAYER_READY]: (payload) => {
    console.log("Player ready:", payload);
  },
  [GameAction.PLAYER_READY_OK]: (payload) => {
    console.log("Player ready OK:", payload);
    useGameStore.getState().setReady(true);
  },
  [GameAction.DUEL_START]: (payload) => {
    const { quest, round } = payload as GameStart;
    useGameStore.getState().setPhase("started");
    useGameStore.getState().setRound(round);
    useGameStore.getState().setQuestion({
      id: crypto.randomUUID(),
      question: quest,
      answers: [],
    });
    // set player scores
  },
  [GameAction.GAME_OVER]: (payload) => {
    useGameStore.getState().setPhase("game_over");
    console.log("Game over:", payload);
  },
  [GameAction.TIMER_UPDATE]: (payload) => {
    useGameStore.getState().setTimer(payload);
  },
  [GameAction.INTERMISSION]: () => {
    useGameStore.getState().setPhase("intermission");
  },
  [GameAction.INTERMISSION_TIMER_UPDATE]: (payload) => {
    useGameStore.getState().setTimer(payload);
  },
  [GameAction.ROOM_DESTROYED]: (payload) => {
    console.warn("Room destroyed:", payload);
    appEventBus.fire("notification", { message: "Room destroyed : Player disconnected", type: "info"  });
    useGameStore.getState().resetGame();
    navigationService.go("/lobby");
  },
};
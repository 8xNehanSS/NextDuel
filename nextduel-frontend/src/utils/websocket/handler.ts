import { GameAction } from "./events";
import { useGameStore } from "../../store/gameStore";
import type { GameStart } from "../../types/game";
import { navigationService } from "../navigationService";

type Handler = (payload: any) => void;

export const handlers: Partial<Record<GameAction, Handler>> = {
  error: (payload) => {
    console.error("Error from server:", payload);
  },
  waiting_queue: () => {
    useGameStore.getState().setPhase("waiting");
  },
  match_found: () => {
    useGameStore.getState().setPhase("matched");
  },
  room_created: (payload) => {
    useGameStore.getState().setRoom(payload);
    navigationService.go("/game");
  },
  room_joined: (payload) => {
    useGameStore.getState().setRoom(payload);
  },
  player_joined: (payload) => {
    console.log("Player joined:", payload);
  },
  player_ready: (payload) => {
    console.log("Player ready:", payload);
  },
  duel_start: (payload) => {
    const { quest, round, players } = payload as GameStart;
    useGameStore.getState().setPhase("started");
    useGameStore.getState().setRound(round);
    useGameStore.getState().setQuestion({
      id: crypto.randomUUID(),
      question: quest,
      answers: [],
    });
    // set player scores
  },
  game_over: (payload) => {
    useGameStore.getState().setPhase("game_over");
    console.log("Game over:", payload);
  },
  timer_update: (payload) => {
    useGameStore.getState().setTimer(payload);
  },
  intermission: (payload) => {
    useGameStore.getState().setPhase("intermission");
  },
  intermission_timer_update: (payload) => {
    useGameStore.getState().setTimer(payload);
  }
};
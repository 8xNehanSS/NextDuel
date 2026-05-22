import { GameAction } from "./events";

type Handler = (payload: any) => void;

export const handlers: Partial<Record<GameAction, Handler>> = {
  join_queue: (payload) => {},
  create_room: (payload) => {},
  player_ready: (payload) => {},
};
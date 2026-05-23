import { create } from "zustand";

type UserStore = {
    username: string | null;
    loggedIn: boolean;
    setUsername: (username: string) => void;
    setLoggedIn: (loggedIn: boolean) => void;
    clearUser: () => void;
};

export const useUserStore = create<UserStore>((set) => ({
    username: null,
    loggedIn: false,
    setUsername: (username) => set({ username }),
    setLoggedIn: (loggedIn) => set({ loggedIn }),
    clearUser: () => set({ username: null, loggedIn: false }),
}));
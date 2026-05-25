import { create } from "zustand";

type UserStore = {
    username: string | null;
    loggedIn: boolean;
    requireRefresh: boolean;
    setUser: (username: string) => void;
    setRequireRefresh: (requireRefresh: boolean) => void;
    clearUser: () => void;
};

export const useUserStore = create<UserStore>((set) => ({
    username: null,
    loggedIn: false,
    requireRefresh: true,
    setUser: (username) => set({ username, loggedIn: true, requireRefresh: false }),
    setRequireRefresh: (requireRefresh) => set({ requireRefresh }),
    clearUser: () => set({ username: null, loggedIn: false, requireRefresh: true }),
}));
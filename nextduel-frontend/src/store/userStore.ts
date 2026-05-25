import { create } from "zustand";

type UserStore = {
    username: string | null;
    loggedIn: boolean;
    authChecked: boolean;
    requireRefresh: boolean;
    setUser: (username: string) => void;
    setAuthChecked: () => void;
    setRequireRefresh: (requireRefresh: boolean) => void;
    clearUser: () => void;
};

export const useUserStore = create<UserStore>((set) => ({
    username: null,
    loggedIn: false,
    authChecked: false,
    requireRefresh: true,
    setUser: (username) => set({ username, loggedIn: true, requireRefresh: false, authChecked: true }),
    setAuthChecked: () => set({ authChecked: true }),
    setRequireRefresh: (requireRefresh) => set({ requireRefresh }),
    clearUser: () => set({ username: null, loggedIn: false, requireRefresh: true, authChecked: true }),
}));
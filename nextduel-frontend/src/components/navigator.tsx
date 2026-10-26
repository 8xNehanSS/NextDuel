import { NavLink } from "react-router";
import { logout } from "../services/api/common";
import { useUserStore } from "../store/userStore";
import { navigationService } from "../utils/navigationService";
import { useGameStore } from "../store/gameStore";
import { socketManager } from "../utils/websocket/socket";
import { appEventBus } from "../utils/eventbus";
import { errorMessage } from "../utils/apiclient";

const links = [
    { to: "/", label: "Home", end: true },
    { to: "/lobby", label: "Play" },
    { to: "/leaderboard", label: "Leaderboard" },
    { to: "/history", label: "History" },
];

const Navigator = () => {
    const loggedIn = useUserStore((state) => state.loggedIn);
    const username = useUserStore((state) => state.username);
    const clearUser = useUserStore((state) => state.clearUser);
    const gameConnected = useGameStore((state) => state.connected);

    async function handleSignOut() {
        try {
            await logout();
        } catch (error) {
            appEventBus.fire("notification", { message: "Sign out failed", description: errorMessage(error), type: "error" });
            return;
        }
        socketManager.disconnect();
        clearUser();
        navigationService.go("/login");
    }

    return (
        <nav className="nav">
            <div className="nav-inner">
                <NavLink to="/" className="nav-brand">
                    <span className="brand-mark">ND</span>
                    <span>NextDuel</span>
                    {gameConnected && <span className="status-indicator" title="Connected to game server" />}
                </NavLink>

                {loggedIn && (
                    <ul className="nav-links">
                        {links.map((link) => (
                            <li key={link.to}>
                                <NavLink to={link.to} end={link.end}>
                                    {link.label}
                                </NavLink>
                            </li>
                        ))}
                    </ul>
                )}

                <div className="nav-actions">
                    {loggedIn ? (
                        <>
                            <span className="user-pill">{username}</span>
                            <button onClick={handleSignOut} className="btn btn-ghost">
                                Sign out
                            </button>
                        </>
                    ) : (
                        <>
                            <button onClick={() => navigationService.go("/login")} className="btn btn-ghost">
                                Sign in
                            </button>
                            <button onClick={() => navigationService.go("/register")} className="btn btn-primary">
                                Register
                            </button>
                        </>
                    )}
                </div>
            </div>
        </nav>
    );
};

export default Navigator;

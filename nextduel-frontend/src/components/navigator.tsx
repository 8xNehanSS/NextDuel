import { Link } from "react-router";
import { logout } from "../services/api/common";
import { useUserStore } from "../store/userStore";
import { navigationService } from "../utils/navigationService";
import { useGameStore } from "../store/gameStore";

const Navigator = () => {
    const userStore = useUserStore();
    const loggedIn = useUserStore((state) => state.loggedIn);
    const username = useUserStore((state) => state.username);

    const gameConnected = useGameStore((state) => state.connected);

    async function handleSignOut() {
        try {
            userStore.clearUser();
            await logout();
            navigationService.go("/login");
        } catch (error) {
            console.error("Failed to sign out:", error);
        }
    }

    function handleSignIn() {
        navigationService.go("/login");
    }

    function handleRegister() {
        navigationService.go("/register");
    }

    return (
        <nav className="nav">
            <div className="nav-inner">
                <div className="nav-brand">
                    <span className="brand-mark">ND</span>
                    <span>NextDuel</span>
                    {gameConnected && <span className="status-indicator" title="Connected to game server" />}
                </div>

                <ul className="nav-links">
                    <li>
                        <Link to="/">Home</Link>
                    </li>
                    <li>
                        <Link to="/lobby">Lobby</Link>
                    </li>
                </ul>

                <div className="nav-actions">
                    {loggedIn && <span className="user-pill">{username}</span>}
                    {loggedIn ? (
                        <button onClick={handleSignOut} className="btn btn-ghost">
                            Sign Out
                        </button>
                    ) : (
                        <>
                            <button onClick={handleSignIn} className="btn btn-ghost">
                                Sign In
                            </button>
                            <button onClick={handleRegister} className="btn btn-primary">
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
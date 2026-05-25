import { Link } from "react-router";
import { logout } from "../services/api/common";
import { useUserStore } from "../store/userStore";
import { navigationService } from "../utils/navigationService";

const Navigator = () => {
    const userStore = useUserStore();
    const loggedIn = useUserStore((state) => state.loggedIn);
    const username = useUserStore((state) => state.username);

    async function handleSignOut() {
        try {
            await logout();
            userStore.clearUser();
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
        <nav className="w-full shadow-md px-6 py-4">
            <div className="max-w-6xl mx-auto flex items-center justify-between">

                {/* Left - Links */}
                <ul className="flex items-center gap-6 text-sm font-medium">
                    <li>
                        <Link
                            to="/"
                            className="hover:underline transition"
                        >
                            Home
                        </Link>
                    </li>

                    <li>
                        <Link
                            to="/lobby"
                            className="hover:underline transition"
                        >
                            Lobby
                        </Link>
                    </li>
                </ul>

                {/* Center - User Info */}
                <div className="text-sm">
                    {loggedIn && (
                        <span>
                            Welcome, {username}
                        </span>
                    )}
                </div>

                {/* Right - Actions */}
                <div>
                    {loggedIn ? (
                        <button
                            onClick={handleSignOut}
                            className="border px-4 py-2 rounded text-sm hover:underline transition"
                        >
                            Sign Out
                        </button>
                    ) : (
                        <>
                            <button
                                onClick={handleSignIn}
                                className="border px-4 py-2 rounded text-sm hover:underline transition"
                            >
                                Sign In
                            </button>
                            <button
                                onClick={handleRegister}
                                className="border px-4 py-2 rounded text-sm hover:underline transition"
                            >
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
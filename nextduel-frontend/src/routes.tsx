import { Link, Navigate, Route, Routes, useLocation, useNavigate } from "react-router";
import { useEffect, type ReactNode } from "react";
import { navigationService } from "./utils/navigationService";
import { useUserStore } from "./store/userStore";
import Homepage from "./pages/home/page";
import Loginpage from "./pages/login/page";
import LobbyPage from "./pages/lobby/page";
import GamePage from "./pages/game/page";
import RegisterPage from "./pages/register/page";
import HistoryPage from "./pages/history/page";
import LeaderboardPage from "./pages/leaderboard/page";

function ProtectedRoute({ children }: { children: ReactNode }) {
    const loggedIn = useUserStore((s) => s.loggedIn);
    const authChecked = useUserStore((s) => s.authChecked);
    const location = useLocation();

    if (!authChecked) return null;
    if (!loggedIn) {
        return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname + location.search)}`} replace />;
    }
    return <>{children}</>;
}

function GuestRoute({ children }: { children: ReactNode }) {
    const loggedIn = useUserStore((s) => s.loggedIn);
    const authChecked = useUserStore((s) => s.authChecked);

    if (!authChecked) return null;
    if (loggedIn) return <Navigate to="/" replace />;
    return <>{children}</>;
}

const NotFound = () => (
    <div className="page-shell">
        <div className="panel empty-state">
            <h1 className="section-title">Page not found</h1>
            <p className="section-subtitle">That page does not exist.</p>
            <div className="hero-actions" style={{ justifyContent: "center" }}>
                <Link to="/" className="btn btn-primary">
                    Go home
                </Link>
            </div>
        </div>
    </div>
);

export default function AppRoutes() {
    const navigate = useNavigate();

    useEffect(() => {
        navigationService.setNavigate(navigate);
    }, [navigate]);

    return (
        <Routes>
            <Route path="/" element={<ProtectedRoute><Homepage /></ProtectedRoute>} />
            <Route path="/login" element={<GuestRoute><Loginpage /></GuestRoute>} />
            <Route path="/register" element={<GuestRoute><RegisterPage /></GuestRoute>} />
            <Route path="/lobby" element={<ProtectedRoute><LobbyPage /></ProtectedRoute>} />
            <Route path="/game" element={<ProtectedRoute><GamePage /></ProtectedRoute>} />
            <Route path="/leaderboard" element={<ProtectedRoute><LeaderboardPage /></ProtectedRoute>} />
            <Route path="/history" element={<ProtectedRoute><HistoryPage /></ProtectedRoute>} />
            <Route path="*" element={<NotFound />} />
        </Routes>
    );
}

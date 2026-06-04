import { Navigate, Route, Routes, useLocation, useNavigate } from "react-router";
import { useEffect } from "react";
import { navigationService } from "./utils/navigationService";
import { useUserStore } from "./store/userStore";
import Homepage from "./pages/home/page";
import Loginpage from "./pages/login/page";
import LobbyPage from "./pages/lobby/page";
import GamePage from "./pages/game/page";
import RegisterPage from "./pages/register/page";
import HistoryPage from "./pages/history/page";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
    const { loggedIn, authChecked } = useUserStore();
    const location = useLocation();

    if (!authChecked) return null;
    if (!loggedIn) {
        return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname)}`} replace />;
    }
    return <>{children}</>;
}

function GuestRoute({ children }: { children: React.ReactNode }) {
    const { loggedIn, authChecked } = useUserStore();

    if (!authChecked) return null;
    if (loggedIn) return <Navigate to="/" replace />;
    return <>{children}</>;
}

const About = () => <div>About Page</div>;
const NotFound = () => <div>Page Not Found</div>;

export default function AppRoutes() {
    const navigate = useNavigate();

    useEffect(() => {
        navigationService.setNavigate(navigate);
    }, [navigate]);

    return (
        <Routes>
            <Route path="/" element={<ProtectedRoute><Homepage /></ProtectedRoute>} />
            <Route path="/about" element={<About />} />
            <Route path="/login" element={<GuestRoute><Loginpage /></GuestRoute>} />
            <Route path="/register" element={<GuestRoute><RegisterPage /></GuestRoute>} />
            <Route path="/lobby" element={<ProtectedRoute><LobbyPage /></ProtectedRoute>} />
            <Route path="/game" element={<ProtectedRoute><GamePage /></ProtectedRoute>} />
            <Route path="/history" element={<ProtectedRoute><HistoryPage /></ProtectedRoute>} />
            <Route path="*" element={<NotFound />} />
        </Routes>
    );
}
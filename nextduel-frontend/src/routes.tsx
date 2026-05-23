import { Route, Routes, useNavigate } from "react-router";
import { useEffect } from "react";
import { navigationService } from "./utils/navigationService";
import Homepage from "./pages/home/page";
import Loginpage from "./pages/login/page";
import LobbyPage from "./pages/lobby/page";
import GamePage from "./pages/game/page";
import RegisterPage from "./pages/register/page";


const About = () => <div>About Page</div>;
const NotFound = () => <div>Page Not Found</div>;

export default function AppRoutes() {
    const navigate = useNavigate();

    useEffect(() => {
        navigationService.setNavigate(navigate);
    }, [navigate]);
    return (
        <Routes>
            <Route path="/" element={<Homepage />} />
            <Route path="/about" element={<About />} />
            <Route path="/login" element={<Loginpage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/lobby" element={<LobbyPage />} />
            <Route path="/game" element={<GamePage />} />

            {/* 404 fallback */}
            <Route path="*" element={<NotFound />} />
        </Routes>
    );
}
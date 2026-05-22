import { Route, Routes } from "react-router";
import Homepage from "./pages/home/page";
import Loginpage from "./pages/login/page";
import LobbyPage from "./pages/lobby/page";

const About = () => <div>About Page</div>;
const NotFound = () => <div>Page Not Found</div>;

export default function AppRoutes() {
    return (
        <Routes>
            <Route path="/" element={<Homepage />} />
            <Route path="/about" element={<About />} />
            <Route path="/login" element={<Loginpage />} />
            <Route path="/lobby" element={<LobbyPage />} />

            {/* 404 fallback */}
            <Route path="*" element={<NotFound />} />
        </Routes>
    );
}
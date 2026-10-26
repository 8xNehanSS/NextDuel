import { BrowserRouter } from "react-router";
import { useEffect } from "react";
import AppRoutes from "./routes";
import Navigator from "./components/navigator";
import Notifier from "./components/notifier";
import ErrorBoundary from "./components/ErrorBoundary";
import { useUserStore } from "./store/userStore";
import { getUser } from "./services/api/common";
import { appEventBus } from "./utils/eventbus";
import { errorMessage } from "./utils/apiclient";
import "./App.css";

function AppContent() {
    const setUser = useUserStore((state) => state.setUser);
    const setAuthChecked = useUserStore((state) => state.setAuthChecked);

    useEffect(() => {
        getUser()
            .then((user) => (user ? setUser(user.username) : setAuthChecked()))
            .catch((error) => {
                appEventBus.fire("notification", {
                    message: "Could not reach the server",
                    description: errorMessage(error),
                    type: "error",
                });
                setAuthChecked();
            });
    }, [setUser, setAuthChecked]);

    return (
        <div className="App">
            <Notifier />
            <Navigator />
            <main>
                <ErrorBoundary>
                    <AppRoutes />
                </ErrorBoundary>
            </main>
        </div>
    );
}

export default function App() {
    return (
        <BrowserRouter>
            <AppContent />
        </BrowserRouter>
    );
}

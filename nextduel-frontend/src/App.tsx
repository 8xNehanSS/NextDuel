import { BrowserRouter } from "react-router";
import AppRoutes from "./routes";
import Navigator from "./components/navigator";
import { useEffect } from "react";
import { useUserStore } from "./store/userStore";
import { getUser } from "./services/api/common";
import Notifier from "./components/notifier";
import { appEventBus } from "./utils/eventbus";
import "./App.css";

function AppContent() {
  const setUser = useUserStore((state) => state.setUser);
  const setAuthChecked = useUserStore((state) => state.setAuthChecked);

  useEffect(() => {
    loadUser();
  }, []);

  async function loadUser() {
    try {
      const user = await getUser();
      if (user && user.username) {
        setUser(user.username);
      } else {
        setAuthChecked();
      }
    } catch (error) {
      appEventBus.fire("notification", { message: "Failed to load user data", type: "error", description: (error as Error).message   });
      setAuthChecked();
    }
  }

  return (
    <div className="App">
      <Notifier />
      <Navigator />
      <AppRoutes />
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
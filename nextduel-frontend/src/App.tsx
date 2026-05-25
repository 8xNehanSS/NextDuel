import { BrowserRouter } from "react-router";
import AppRoutes from "./routes";
import Navigator from "./components/navigator";
import { useEffect } from "react";
import { useUserStore } from "./store/userStore";
import { getUser } from "./services/api/common";

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
    } catch {
      setAuthChecked();
    }
  }

  return (
    <div className="App">
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
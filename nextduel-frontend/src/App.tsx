import { BrowserRouter } from "react-router";
import AppRoutes from "./routes";
import Navigator from "./components/navigator";
import { useEffect } from "react";
import { useUserStore } from "./store/userStore";
import { getUser } from "./services/api/common";
  
function AppContent() {
  const userStore = useUserStore();

  useEffect(() => {
    loadUser();
  }, []);

  async function loadUser() {
    try {
      const user = await getUser();
      if (user && user.username) {
        userStore.setUser(user.username);
      }
    } catch (error) {
      console.error("Failed to load user:", error);
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
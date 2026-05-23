import { BrowserRouter } from "react-router";
import AppRoutes from "./routes";
import Navigator from "./components/navigator";
import { useEffect } from "react";
import { useUserStore } from "./store/userStore";
import { getUser } from "./services/api/common";


export default function App() {

  const userStore = useUserStore();

  useEffect(() => {
    loadUser();
  }, []);

  async function loadUser() {
    try {
      const user = await getUser();
      userStore.setUsername(user.username);
      userStore.setLoggedIn(true);
    } catch (error) {
      console.error("Failed to load user:", error);
    }
  }

  return (
    <BrowserRouter>
      <div className="App">
        <Navigator />
        <AppRoutes />
      </div>
    </BrowserRouter>
  );
}
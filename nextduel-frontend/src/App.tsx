import { BrowserRouter } from "react-router";
import AppRoutes from "./routes";
import Navigator from "./components/navigator";

export default function App() {
  return (
    <BrowserRouter>
      <div className="App">
        <Navigator />
        <AppRoutes />
      </div>
    </BrowserRouter>
  );
}
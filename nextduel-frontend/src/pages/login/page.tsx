import { useState } from "react";
import { login, type LoginRequest } from "../../services/api/common";

const loginpage = () => {

    const [loginData, setLoginData] = useState<LoginRequest | null>(null);

    async function handleLogin() {
        if (loginData) {
            await login(loginData);
        }
    }

    return (
        <div>
            <h1>Login Page</h1>
            <p>Please enter your credentials to log in.</p>
            <input
                type="text"
                placeholder="Username"
                onChange={(e) => setLoginData({ ...loginData, username: e.target.value } as LoginRequest)}
            />
            <input
                type="password"
                placeholder="Password"
                onChange={(e) => setLoginData({ ...loginData, password: e.target.value } as LoginRequest)}
            />
            <button onClick={handleLogin}>Login</button>
        </div>
    );
}

export default loginpage;
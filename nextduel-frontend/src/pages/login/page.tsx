import { useState } from "react";
import { login, type LoginRequest } from "../../services/api/common";

const LoginPage = () => {
    const [loginData, setLoginData] = useState<LoginRequest>({
        username: "",
        password: "",
    });
    const [error, setError] = useState<string | null>(null);

    async function handleLogin() {
        if (!loginData.username || !loginData.password) return;
        setError(null);
        try {
            await login(loginData);
        } catch {
            setError("Invalid username or password");
        }
    }

    return (
        <div className="flex items-center justify-center px-4 mt-4">
            <div className="w-full max-w-md bg-white shadow-xl rounded-2xl p-8">
                <p className="text-sm text-gray-500 text-center mt-2">
                    Please enter your credentials to log in
                </p>

                <div className="mt-6 space-y-4">

                    <input
                        type="text"
                        placeholder="Username"
                        value={loginData.username}
                        onChange={(e) =>
                            setLoginData((prev) => ({
                                ...prev,
                                username: e.target.value,
                            }))
                        }
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg 
                                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />

                    <input
                        type="password"
                        placeholder="Password"
                        value={loginData.password}
                        onChange={(e) =>
                            setLoginData((prev) => ({
                                ...prev,
                                password: e.target.value,
                            }))
                        }
                        className="w-full px-4 py-3 border border-gray-300 rounded-lg 
                                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />

                    {error && (
                        <p className="text-red-500 text-sm text-center">{error}</p>
                    )}

                    <button
                        onClick={handleLogin}
                        className="w-full bg-blue-600 text-white py-3 rounded-lg font-semibold
                                   hover:bg-blue-700 active:scale-[0.98] transition"
                    >
                        Login
                    </button>
                </div>
            </div>
        </div>
    );
};

export default LoginPage;
import { useState } from "react";
import { register, type RegisterRequest } from "../../services/api/common";

const RegisterPage = () => {
    const [userInformation, setUserInformation] = useState<RegisterRequest>({
        username: "",
        password: "",
        email: "",
    });

    async function handleRegister() {
        try {
            await register(userInformation);
        } catch (error) {
            console.log("Registration failed:", error);
        }
    }

    return (
        <div className="min-h-screen flex items-center justify-center px-4 bg-gray-100">
            <div className="w-full max-w-md bg-white shadow-xl rounded-2xl p-8">

                <h1 className="text-2xl font-bold text-center">
                    Create Account
                </h1>

                <p className="text-sm text-center mt-2 text-gray-500">
                    Register to get started
                </p>

                <div className="mt-6 space-y-4">

                    {/* Username */}
                    <input
                        type="text"
                        placeholder="Username"
                        value={userInformation.username}
                        onChange={(e) =>
                            setUserInformation((prev) => ({
                                ...prev,
                                username: e.target.value,
                            }))
                        }
                        className="w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />

                    {/* Email */}
                    <input
                        type="email"
                        placeholder="Email"
                        value={userInformation.email}
                        onChange={(e) =>
                            setUserInformation((prev) => ({
                                ...prev,
                                email: e.target.value,
                            }))
                        }
                        className="w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />

                    {/* Password */}
                    <input
                        type="password"
                        placeholder="Password"
                        value={userInformation.password}
                        onChange={(e) =>
                            setUserInformation((prev) => ({
                                ...prev,
                                password: e.target.value,
                            }))
                        }
                        className="w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />

                    {/* Button */}
                    <button
                        onClick={handleRegister}
                        className="w-full border py-3 rounded-lg font-medium hover:underline transition"
                    >
                        Register
                    </button>

                </div>
            </div>
        </div>
    );
};

export default RegisterPage;
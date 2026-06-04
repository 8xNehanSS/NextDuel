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

    const handleSubmit = async (e: React.SubmitEvent) => {
        e.preventDefault();
        await handleRegister();
    };

    return (
        <div className="flex items-center justify-center px-4 pt-6">
            <div className="w-full max-w-md bg-white shadow-xl rounded-2xl p-8">

                <h1 className="text-2xl font-bold text-center">
                    Create Account
                </h1>

                <p className="text-sm text-center mt-2 text-gray-500">
                    Register to get started
                </p>
                <form onSubmit={handleSubmit}>
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
                            type="submit"
                            className="w-full border py-3 rounded-lg font-medium hover:underline transition"
                        >
                            Register
                        </button>

                    </div>
                </form>
            </div>
        </div>
    );
};

export default RegisterPage;
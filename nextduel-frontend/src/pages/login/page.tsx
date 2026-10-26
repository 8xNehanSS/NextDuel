import { useState, type FormEvent } from "react";
import { Link, useSearchParams } from "react-router";
import { login, type LoginRequest } from "../../services/api/common";
import { useUserStore } from "../../store/userStore";
import { navigationService } from "../../utils/navigationService";
import { errorMessage } from "../../utils/apiclient";
import { safeRedirect } from "../../utils/redirect";

const LoginPage = () => {
    const [form, setForm] = useState<LoginRequest>({ username: "", password: "" });
    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);
    const [params] = useSearchParams();
    const setUser = useUserStore((s) => s.setUser);

    async function handleSubmit(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        if (!form.username.trim() || !form.password) {
            setError("Please enter both username and password");
            return;
        }
        setError(null);
        setSubmitting(true);
        try {
            const user = await login({ ...form, username: form.username.trim() });
            setUser(user.username);
            navigationService.go(safeRedirect(params.get("redirect")));
        } catch (err) {
            setError(errorMessage(err, "Login failed"));
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className="page-shell auth-shell">
            <form className="panel auth-card float-in" onSubmit={handleSubmit} noValidate>
                <h1 className="section-title">Welcome back</h1>
                <p className="section-subtitle">Sign in to jump into a duel.</p>

                <label className="form-field">
                    <span className="form-label">Username</span>
                    <input
                        className="input-field"
                        autoComplete="username"
                        autoFocus
                        value={form.username}
                        onChange={(e) => setForm((prev) => ({ ...prev, username: e.target.value }))}
                    />
                </label>

                <label className="form-field">
                    <span className="form-label">Password</span>
                    <input
                        className="input-field"
                        type="password"
                        autoComplete="current-password"
                        value={form.password}
                        onChange={(e) => setForm((prev) => ({ ...prev, password: e.target.value }))}
                    />
                </label>

                {error && (
                    <p className="form-error" role="alert">
                        {error}
                    </p>
                )}

                <button type="submit" className="btn btn-primary btn-block" disabled={submitting}>
                    {submitting ? "Signing in..." : "Sign in"}
                </button>

                <p className="auth-footer">
                    New to NextDuel? <Link to="/register">Create an account</Link>
                </p>
            </form>
        </div>
    );
};

export default LoginPage;

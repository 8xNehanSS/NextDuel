import { useState, type FormEvent } from "react";
import { Link } from "react-router";
import { register, type RegisterRequest } from "../../services/api/common";
import { navigationService } from "../../utils/navigationService";
import { appEventBus } from "../../utils/eventbus";
import { errorMessage } from "../../utils/apiclient";

type Errors = Partial<Record<keyof RegisterRequest | "form", string>>;

// Mirrors server-side validation in nextduel-svr/routes/register/post.go.
function validate(form: RegisterRequest): Errors {
    const errors: Errors = {};
    if (!/^[A-Za-z0-9_]{3,20}$/.test(form.username.trim())) {
        errors.username = "3-20 characters: letters, numbers or underscores.";
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) {
        errors.email = "Enter a valid email address.";
    }
    if (form.password.length < 8) {
        errors.password = "At least 8 characters.";
    }
    return errors;
}

const RegisterPage = () => {
    const [form, setForm] = useState<RegisterRequest>({ username: "", password: "", email: "" });
    const [errors, setErrors] = useState<Errors>({});
    const [submitting, setSubmitting] = useState(false);

    const update = (key: keyof RegisterRequest) => (e: React.ChangeEvent<HTMLInputElement>) =>
        setForm((prev) => ({ ...prev, [key]: e.target.value }));

    async function handleSubmit(e: FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const found = validate(form);
        setErrors(found);
        if (Object.keys(found).length > 0) return;

        setSubmitting(true);
        try {
            await register({ ...form, username: form.username.trim(), email: form.email.trim() });
            appEventBus.fire("notification", { message: "Account created. Please sign in.", type: "success" });
            navigationService.go("/login");
        } catch (err) {
            setErrors({ form: errorMessage(err, "Registration failed") });
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className="page-shell auth-shell">
            <form className="panel auth-card float-in" onSubmit={handleSubmit} noValidate>
                <h1 className="section-title">Create account</h1>
                <p className="section-subtitle">Pick a name your rivals will remember.</p>

                <label className="form-field">
                    <span className="form-label">Username</span>
                    <input
                        className="input-field"
                        autoComplete="username"
                        autoFocus
                        value={form.username}
                        onChange={update("username")}
                        aria-invalid={!!errors.username}
                    />
                    {errors.username && <span className="field-error">{errors.username}</span>}
                </label>

                <label className="form-field">
                    <span className="form-label">Email</span>
                    <input
                        className="input-field"
                        type="email"
                        autoComplete="email"
                        value={form.email}
                        onChange={update("email")}
                        aria-invalid={!!errors.email}
                    />
                    {errors.email && <span className="field-error">{errors.email}</span>}
                </label>

                <label className="form-field">
                    <span className="form-label">Password</span>
                    <input
                        className="input-field"
                        type="password"
                        autoComplete="new-password"
                        value={form.password}
                        onChange={update("password")}
                        aria-invalid={!!errors.password}
                    />
                    {errors.password && <span className="field-error">{errors.password}</span>}
                </label>

                {errors.form && (
                    <p className="form-error" role="alert">
                        {errors.form}
                    </p>
                )}

                <button type="submit" className="btn btn-primary btn-block" disabled={submitting}>
                    {submitting ? "Creating account..." : "Create account"}
                </button>

                <p className="auth-footer">
                    Already have an account? <Link to="/login">Sign in</Link>
                </p>
            </form>
        </div>
    );
};

export default RegisterPage;

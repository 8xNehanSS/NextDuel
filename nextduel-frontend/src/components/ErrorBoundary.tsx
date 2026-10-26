import { Component, type ErrorInfo, type ReactNode } from "react";

type Props = { children: ReactNode };
type State = { error: Error | null };

/** Catches render errors so one broken page does not blank the whole app. */
export default class ErrorBoundary extends Component<Props, State> {
    state: State = { error: null };

    static getDerivedStateFromError(error: Error): State {
        return { error };
    }

    componentDidCatch(error: Error, info: ErrorInfo) {
        console.error("Unhandled render error", error, info.componentStack);
    }

    render() {
        if (!this.state.error) {
            return this.props.children;
        }
        return (
            <div className="page-shell">
                <div className="panel empty-state">
                    <h1 className="section-title">Something went wrong</h1>
                    <p className="section-subtitle">An unexpected error occurred. Reloading usually fixes it.</p>
                    <div className="hero-actions" style={{ justifyContent: "center" }}>
                        <button className="btn btn-primary" onClick={() => window.location.assign("/")}>
                            Reload NextDuel
                        </button>
                    </div>
                </div>
            </div>
        );
    }
}

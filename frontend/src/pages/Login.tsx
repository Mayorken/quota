import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth";
import GoogleButton from "../components/GoogleButton";

export default function Login() {
  const { login, signup, loginWithGoogle } = useAuth();
  const navigate = useNavigate();
  const [mode, setMode] = useState<"login" | "signup">("login");
  const [orgName, setOrgName] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      if (mode === "login") {
        await login(email, password);
      } else {
        await signup(orgName, name, email, password);
      }
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setBusy(false);
    }
  }

  function fillDemo() {
    setMode("login");
    setEmail("demo@quota.app");
    setPassword("password123");
  }

  async function handleGoogle(credential: string) {
    setError("");
    setBusy(true);
    try {
      await loginWithGoogle(credential, mode === "signup" ? orgName : undefined);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Google sign-in failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="auth-wrap">
      <div className="auth-card">
        <div className="auth-brand">
          <span className="brand-mark big">Q</span>
          <h1>Quota</h1>
          <p className="muted">
            Commission tracking your sales team actually trusts.
          </p>
        </div>

        <div className="tabs">
          <button
            className={mode === "login" ? "tab active" : "tab"}
            onClick={() => setMode("login")}
          >
            Sign in
          </button>
          <button
            className={mode === "signup" ? "tab active" : "tab"}
            onClick={() => setMode("signup")}
          >
            Create account
          </button>
        </div>

        <form onSubmit={submit}>
          {mode === "signup" && (
            <>
              <label>Company name</label>
              <input
                value={orgName}
                onChange={(e) => setOrgName(e.target.value)}
                placeholder="Acme Inc."
                required
              />
              <label>Your name</label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Jane Smith"
                required
              />
            </>
          )}
          <label>Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            required
          />
          <label>Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Min. 8 characters"
            required
            minLength={8}
          />

          {error && <div className="error">{error}</div>}

          <button className="btn primary full" type="submit" disabled={busy}>
            {busy
              ? "Please wait..."
              : mode === "login"
                ? "Sign in"
                : "Create account"}
          </button>
        </form>

        <GoogleButton onCredential={handleGoogle} />

        <button className="btn link full" onClick={fillDemo} type="button">
          Use demo account
        </button>
      </div>
    </div>
  );
}

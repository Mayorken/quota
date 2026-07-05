import { Navigate, Route, Routes, Link, useLocation } from "react-router-dom";
import { useAuth, isManager } from "./auth";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import CompPlans from "./pages/CompPlans";
import Deals from "./pages/Deals";
import Team from "./pages/Team";
import RepDetail from "./pages/RepDetail";

function Nav() {
  const { user, logout } = useAuth();
  const loc = useLocation();
  if (!user) return null;

  const link = (to: string, label: string) => (
    <Link to={to} className={loc.pathname === to ? "nav-link active" : "nav-link"}>
      {label}
    </Link>
  );

  return (
    <nav className="topnav">
      <div className="brand">
        <span className="brand-mark">Q</span> Quota
      </div>
      <div className="nav-links">
        {link("/", "Dashboard")}
        {isManager(user) && link("/deals", "Deals")}
        {isManager(user) && link("/comp-plans", "Comp Plans")}
        {isManager(user) && link("/team", "Team")}
      </div>
      <div className="nav-user">
        <span className="who">
          {user.name} <span className="role-badge">{user.role}</span>
        </span>
        <button className="btn ghost" onClick={logout}>
          Sign out
        </button>
      </div>
    </nav>
  );
}

function RequireAuth({ children }: { children: JSX.Element }) {
  const { user, loading } = useAuth();
  if (loading) return <div className="center-screen">Loading…</div>;
  if (!user) return <Navigate to="/login" replace />;
  return children;
}

function RequireManager({ children }: { children: JSX.Element }) {
  const { user } = useAuth();
  if (!isManager(user)) return <Navigate to="/" replace />;
  return children;
}

export default function App() {
  const { user, loading } = useAuth();

  return (
    <div className="app">
      <Nav />
      <main className="content">
        <Routes>
          <Route
            path="/login"
            element={user && !loading ? <Navigate to="/" replace /> : <Login />}
          />
          <Route
            path="/"
            element={
              <RequireAuth>
                <Dashboard />
              </RequireAuth>
            }
          />
          <Route
            path="/reps/:id"
            element={
              <RequireAuth>
                <RepDetail />
              </RequireAuth>
            }
          />
          <Route
            path="/deals"
            element={
              <RequireAuth>
                <RequireManager>
                  <Deals />
                </RequireManager>
              </RequireAuth>
            }
          />
          <Route
            path="/comp-plans"
            element={
              <RequireAuth>
                <RequireManager>
                  <CompPlans />
                </RequireManager>
              </RequireAuth>
            }
          />
          <Route
            path="/team"
            element={
              <RequireAuth>
                <RequireManager>
                  <Team />
                </RequireManager>
              </RequireAuth>
            }
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}

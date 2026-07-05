import { Navigate, Route, Routes, Link, useLocation } from "react-router-dom";
import { useAuth, isManager } from "./auth";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import CompPlans from "./pages/CompPlans";
import Deals from "./pages/Deals";
import Team from "./pages/Team";
import RepDetail from "./pages/RepDetail";

/* ---- SVG icons (inline to avoid a dependency) ---- */
const Icons = {
  dashboard: (
    <svg viewBox="0 0 20 20" fill="currentColor" className="nav-icon">
      <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 6a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zm10 0a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
    </svg>
  ),
  deals: (
    <svg viewBox="0 0 20 20" fill="currentColor" className="nav-icon">
      <path fillRule="evenodd" d="M4 4a2 2 0 00-2 2v4a2 2 0 002 2V6h10a2 2 0 00-2-2H4zm2 6a2 2 0 012-2h8a2 2 0 012 2v4a2 2 0 01-2 2H8a2 2 0 01-2-2v-4zm6 4a2 2 0 100-4 2 2 0 000 4z" clipRule="evenodd" />
    </svg>
  ),
  plans: (
    <svg viewBox="0 0 20 20" fill="currentColor" className="nav-icon">
      <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" />
      <path fillRule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clipRule="evenodd" />
    </svg>
  ),
  team: (
    <svg viewBox="0 0 20 20" fill="currentColor" className="nav-icon">
      <path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zm8 0a3 3 0 11-6 0 3 3 0 016 0zm-4.07 11c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z" />
    </svg>
  ),
  signout: (
    <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14">
      <path fillRule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 9.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 9H7a1 1 0 100 2h7.586l-1.293 1.293z" clipRule="evenodd" />
    </svg>
  ),
};

const pageTitle: Record<string, string> = {
  "/": "Dashboard",
  "/deals": "Deals",
  "/comp-plans": "Comp Plans",
  "/team": "Team",
};

function Sidebar() {
  const { user, logout } = useAuth();
  const loc = useLocation();
  if (!user) return null;

  const link = (to: string, label: string, icon: JSX.Element) => (
    <Link
      to={to}
      className={`sidebar-link${loc.pathname === to ? " active" : ""}`}
    >
      {icon}
      {label}
    </Link>
  );

  const initials = user.name
    ? user.name
        .split(" ")
        .map((w) => w[0])
        .join("")
        .slice(0, 2)
        .toUpperCase()
    : user.email[0].toUpperCase();

  return (
    <aside className="sidebar">
      <div className="sidebar-brand">
        <span className="brand-mark">Q</span>
        <span className="brand-text">Quota</span>
      </div>

      <nav className="sidebar-nav">
        <div className="sidebar-section">Overview</div>
        {link("/", "Dashboard", Icons.dashboard)}

        {isManager(user) && (
          <>
            <div className="sidebar-section">Manage</div>
            {link("/deals", "Deals", Icons.deals)}
            {link("/comp-plans", "Comp Plans", Icons.plans)}
            {link("/team", "Team", Icons.team)}
          </>
        )}
      </nav>

      <div className="sidebar-footer">
        <div className="sidebar-user">
          <div className="sidebar-avatar">{initials}</div>
          <div className="sidebar-user-info">
            <div className="sidebar-user-name">{user.name}</div>
            <div className="sidebar-user-role">{user.role}</div>
          </div>
        </div>
        <button className="sidebar-signout" onClick={logout}>
          {Icons.signout} Sign out
        </button>
      </div>
    </aside>
  );
}

function Topbar() {
  const loc = useLocation();
  const repMatch = loc.pathname.match(/^\/reps\//);
  const title = repMatch ? "Rep Detail" : pageTitle[loc.pathname] || "";
  return (
    <header className="topbar">
      <span className="topbar-title">{title}</span>
      <div className="topbar-actions" />
    </header>
  );
}

function RequireAuth({ children }: { children: JSX.Element }) {
  const { user, loading } = useAuth();
  if (loading) return <div className="center-screen">Loading...</div>;
  if (!user) return <Navigate to="/login" replace />;
  return children;
}

function RequireManager({ children }: { children: JSX.Element }) {
  const { user } = useAuth();
  if (!isManager(user)) return <Navigate to="/" replace />;
  return children;
}

function AuthenticatedLayout({ children }: { children: JSX.Element }) {
  return (
    <>
      <Sidebar />
      <div className="main-area">
        <Topbar />
        <main className="content">{children}</main>
      </div>
    </>
  );
}

export default function App() {
  const { user, loading } = useAuth();

  return (
    <div className="app">
      <Routes>
        <Route
          path="/login"
          element={user && !loading ? <Navigate to="/" replace /> : <Login />}
        />
        <Route
          path="/"
          element={
            <RequireAuth>
              <AuthenticatedLayout>
                <Dashboard />
              </AuthenticatedLayout>
            </RequireAuth>
          }
        />
        <Route
          path="/reps/:id"
          element={
            <RequireAuth>
              <AuthenticatedLayout>
                <RepDetail />
              </AuthenticatedLayout>
            </RequireAuth>
          }
        />
        <Route
          path="/deals"
          element={
            <RequireAuth>
              <RequireManager>
                <AuthenticatedLayout>
                  <Deals />
                </AuthenticatedLayout>
              </RequireManager>
            </RequireAuth>
          }
        />
        <Route
          path="/comp-plans"
          element={
            <RequireAuth>
              <RequireManager>
                <AuthenticatedLayout>
                  <CompPlans />
                </AuthenticatedLayout>
              </RequireManager>
            </RequireAuth>
          }
        />
        <Route
          path="/team"
          element={
            <RequireAuth>
              <RequireManager>
                <AuthenticatedLayout>
                  <Team />
                </AuthenticatedLayout>
              </RequireManager>
            </RequireAuth>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
}

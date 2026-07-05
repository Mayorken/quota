import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { api, clearToken, getToken, setToken, User } from "./api";

interface AuthState {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (orgName: string, name: string, email: string, password: string) => Promise<void>;
  loginWithGoogle: (credential: string, orgName?: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

interface AuthResponse {
  token: string;
  user: User;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Restore session on load if a token exists.
    if (!getToken()) {
      setLoading(false);
      return;
    }
    api
      .get<User>("/auth/me")
      .then(setUser)
      .catch(() => clearToken())
      .finally(() => setLoading(false));
  }, []);

  async function login(email: string, password: string) {
    const res = await api.post<AuthResponse>("/auth/login", { email, password });
    setToken(res.token);
    setUser(res.user);
  }

  async function signup(orgName: string, name: string, email: string, password: string) {
    const res = await api.post<AuthResponse>("/auth/signup", {
      org_name: orgName,
      name,
      email,
      password,
    });
    setToken(res.token);
    setUser(res.user);
  }

  async function loginWithGoogle(credential: string, orgName?: string) {
    const res = await api.post<AuthResponse>("/auth/google", {
      credential,
      org_name: orgName,
    });
    setToken(res.token);
    setUser(res.user);
  }

  function logout() {
    clearToken();
    setUser(null);
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, signup, loginWithGoogle, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

export function isManager(user: User | null): boolean {
  return user?.role === "manager" || user?.role === "admin";
}

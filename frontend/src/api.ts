// Thin fetch wrapper that attaches the JWT and parses JSON errors.

const TOKEN_KEY = "quota_token";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  const token = getToken();
  if (token) headers["Authorization"] = `Bearer ${token}`;
  if (body !== undefined) headers["Content-Type"] = "application/json";

  let res: Response;
  try {
    res = await fetch(`/api${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch {
    // Network-level failure (backend not running / unreachable).
    throw new ApiError(0, "Can't reach the server. Is the backend running on port 8080?");
  }

  if (!res.ok) {
    let msg = res.statusText;
    try {
      const data = await res.json();
      msg = data.error || msg;
    } catch {
      /* non-JSON error (e.g. the dev proxy couldn't reach the backend) */
    }
    // A 5xx with no JSON body almost always means the backend is down.
    if (res.status >= 500 && (msg === res.statusText || !msg)) {
      msg = "Can't reach the server. Start the backend (cd backend && go run ./cmd/server).";
    }
    throw new ApiError(res.status, msg);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>("GET", path),
  post: <T>(path: string, body?: unknown) => request<T>("POST", path, body),
  put: <T>(path: string, body?: unknown) => request<T>("PUT", path, body),
  del: <T>(path: string) => request<T>("DELETE", path),
};

// ---- Types shared with the backend ----

export interface User {
  id: string;
  org_id: string;
  email: string;
  name: string;
  role: "rep" | "manager" | "admin";
}

export interface Tier {
  from_pct: number;
  to_pct: number;
  rate_pct: number;
}

export interface CompPlan {
  id: string;
  name: string;
  period_type: "monthly" | "quarterly" | "annual";
  quota_amount: number;
  base_salary: number;
  tiers: Tier[];
  type_multipliers: Record<string, number> | null;
  effective_date: string;
}

export interface Deal {
  id: string;
  rep_id: string;
  amount: number;
  deal_type: string;
  close_date: string;
  rep?: User;
}

export interface BreakdownLine {
  label: string;
  from_pct: number;
  to_pct: number;
  rate_pct: number;
  revenue_in_band: number;
  commission: number;
}

export interface Breakdown {
  quota: number;
  attainment_pct: number;
  total_revenue: number;
  credited_revenue: number;
  deal_count: number;
  lines: BreakdownLine[];
  type_multipliers?: Record<string, number>;
  total_commission: number;
}

export interface RepResult {
  rep_id: string;
  rep_name: string;
  rep_email: string;
  comp_plan_id: string;
  comp_plan_name: string;
  period_start: string;
  period_end: string;
  quota: number;
  attained: number;
  attainment_pct: number;
  commission_owed: number;
  breakdown: Breakdown;
  has_plan: boolean;
}

export interface DashboardResponse {
  as_of: string;
  results: RepResult[];
}

export interface Assignment {
  id: string;
  user_id: string;
  comp_plan_id: string;
  comp_plan?: CompPlan;
  user?: User;
}

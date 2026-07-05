import { useEffect, useState } from "react";
import { api, User, CompPlan, Assignment } from "../api";

export default function Team() {
  const [users, setUsers] = useState<User[]>([]);
  const [plans, setPlans] = useState<CompPlan[]>([]);
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [error, setError] = useState("");

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"rep" | "manager">("rep");

  function reload() {
    api.get<User[]>("/users").then(setUsers).catch((e) => setError(e.message));
    api.get<CompPlan[]>("/comp-plans").then(setPlans);
    api.get<Assignment[]>("/comp-plans/assignments").then(setAssignments);
  }
  useEffect(reload, []);

  async function addUser(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await api.post("/users", { name, email, password, role });
      setName("");
      setEmail("");
      setPassword("");
      reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed");
    }
  }

  async function assign(userId: string, compPlanId: string) {
    if (!compPlanId) return;
    await api.post("/comp-plans/assign", { user_id: userId, comp_plan_id: compPlanId });
    reload();
  }

  function currentPlan(userId: string): string {
    const a = assignments.find((x) => x.user_id === userId);
    return a?.comp_plan?.name ?? "\u2014";
  }

  return (
    <div>
      <div className="page-head">
        <div>
          <h2>Team</h2>
          <p className="muted">{users.length} member{users.length !== 1 ? "s" : ""}</p>
        </div>
      </div>

      {error && <div className="error">{error}</div>}

      <form className="card form-inline" onSubmit={addUser}>
        <div className="field">
          <label>Name</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Jane Smith"
            required
          />
        </div>
        <div className="field">
          <label>Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="jane@company.com"
            required
          />
        </div>
        <div className="field">
          <label>Temp password</label>
          <input
            type="text"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Min. 8 characters"
            minLength={8}
            required
          />
        </div>
        <div className="field">
          <label>Role</label>
          <select value={role} onChange={(e) => setRole(e.target.value as "rep" | "manager")}>
            <option value="rep">Rep</option>
            <option value="manager">Manager</option>
          </select>
        </div>
        <button className="btn primary" type="submit">
          <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14">
            <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z" />
          </svg>
          Add Member
        </button>
      </form>

      <table className="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
            <th>Comp Plan</th>
            <th>Assign Plan</th>
          </tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id}>
              <td style={{ fontWeight: 500 }}>{u.name}</td>
              <td className="muted">{u.email}</td>
              <td>
                <span className={`role-badge${u.role !== "rep" ? ` role-badge-${u.role}` : ""}`}>
                  {u.role}
                </span>
              </td>
              <td style={{ fontWeight: 500 }}>{currentPlan(u.id)}</td>
              <td>
                {u.role === "rep" ? (
                  <select defaultValue="" onChange={(e) => assign(u.id, e.target.value)} style={{ width: "auto", minWidth: 140 }}>
                    <option value="" disabled>
                      Assign...
                    </option>
                    {plans.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <span className="muted">\u2014</span>
                )}
              </td>
            </tr>
          ))}
          {users.length === 0 && (
            <tr>
              <td colSpan={5} className="muted center" style={{ padding: "32px 16px" }}>
                No team members yet.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

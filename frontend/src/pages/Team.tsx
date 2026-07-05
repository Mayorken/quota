import { useEffect, useState } from "react";
import { api, User, CompPlan, Assignment } from "../api";

export default function Team() {
  const [users, setUsers] = useState<User[]>([]);
  const [plans, setPlans] = useState<CompPlan[]>([]);
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [error, setError] = useState("");

  // Add-user form.
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
    return a?.comp_plan?.name ?? "—";
  }

  return (
    <div>
      <div className="page-head">
        <h2>Team</h2>
      </div>

      {error && <div className="error">{error}</div>}

      <form className="card form-inline" onSubmit={addUser}>
        <div className="field">
          <label>Name</label>
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="field">
          <label>Email</label>
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </div>
        <div className="field">
          <label>Temp password</label>
          <input
            type="text"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={8}
            required
          />
        </div>
        <div className="field">
          <label>Role</label>
          <select value={role} onChange={(e) => setRole(e.target.value as any)}>
            <option value="rep">rep</option>
            <option value="manager">manager</option>
          </select>
        </div>
        <button className="btn primary" type="submit">
          Add member
        </button>
      </form>

      <table className="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
            <th>Comp plan</th>
            <th>Assign plan</th>
          </tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id}>
              <td>{u.name}</td>
              <td>{u.email}</td>
              <td>
                <span className="role-badge">{u.role}</span>
              </td>
              <td>{currentPlan(u.id)}</td>
              <td>
                {u.role === "rep" ? (
                  <select defaultValue="" onChange={(e) => assign(u.id, e.target.value)}>
                    <option value="" disabled>
                      Assign…
                    </option>
                    {plans.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <span className="muted">—</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

import { useEffect, useState } from "react";
import { api, CompPlan, Tier } from "../api";
import { money } from "../format";

const blankTier = (): Tier => ({ from_pct: 0, to_pct: 0, rate_pct: 0.05 });

export default function CompPlans() {
  const [plans, setPlans] = useState<CompPlan[]>([]);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);

  // Form state.
  const [name, setName] = useState("");
  const [period, setPeriod] = useState<"monthly" | "quarterly" | "annual">("monthly");
  const [quota, setQuota] = useState("100000");
  const [base, setBase] = useState("0");
  const [tiers, setTiers] = useState<Tier[]>([
    { from_pct: 0, to_pct: 1, rate_pct: 0.05 },
    { from_pct: 1, to_pct: 0, rate_pct: 0.08 },
  ]);

  function reload() {
    api.get<CompPlan[]>("/comp-plans").then(setPlans).catch((e) => setError(e.message));
  }
  useEffect(reload, []);

  function updateTier(i: number, key: keyof Tier, val: number) {
    setTiers((t) => t.map((tier, idx) => (idx === i ? { ...tier, [key]: val } : tier)));
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await api.post("/comp-plans", {
        name,
        period_type: period,
        quota_amount: parseFloat(quota),
        base_salary: parseFloat(base) || 0,
        tiers,
      });
      setShowForm(false);
      setName("");
      reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed");
    }
  }

  async function remove(id: string) {
    if (!confirm("Delete this comp plan?")) return;
    await api.del(`/comp-plans/${id}`);
    reload();
  }

  return (
    <div>
      <div className="page-head">
        <h2>Comp plans</h2>
        <button className="btn primary" onClick={() => setShowForm((s) => !s)}>
          {showForm ? "Cancel" : "+ New plan"}
        </button>
      </div>

      {error && <div className="error">{error}</div>}

      {showForm && (
        <form className="card" onSubmit={submit}>
          <div className="grid-2">
            <div className="field">
              <label>Plan name</label>
              <input value={name} onChange={(e) => setName(e.target.value)} required />
            </div>
            <div className="field">
              <label>Period</label>
              <select value={period} onChange={(e) => setPeriod(e.target.value as any)}>
                <option value="monthly">Monthly</option>
                <option value="quarterly">Quarterly</option>
                <option value="annual">Annual</option>
              </select>
            </div>
            <div className="field">
              <label>Quota ($)</label>
              <input
                type="number"
                value={quota}
                onChange={(e) => setQuota(e.target.value)}
                required
              />
            </div>
            <div className="field">
              <label>Base / draw ($, optional)</label>
              <input type="number" value={base} onChange={(e) => setBase(e.target.value)} />
            </div>
          </div>

          <h4>Commission tiers</h4>
          <p className="muted small">
            Bands are % of quota. Rate applies to revenue in that band. Set the top tier's
            "to %" to 0 for uncapped (accelerator).
          </p>
          <table className="tier-table">
            <thead>
              <tr>
                <th>From (% quota)</th>
                <th>To (% quota, 0 = ∞)</th>
                <th>Rate (%)</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {tiers.map((t, i) => (
                <tr key={i}>
                  <td>
                    <input
                      type="number"
                      value={t.from_pct * 100}
                      onChange={(e) => updateTier(i, "from_pct", Number(e.target.value) / 100)}
                    />
                  </td>
                  <td>
                    <input
                      type="number"
                      value={t.to_pct * 100}
                      onChange={(e) => updateTier(i, "to_pct", Number(e.target.value) / 100)}
                    />
                  </td>
                  <td>
                    <input
                      type="number"
                      step="0.1"
                      value={t.rate_pct * 100}
                      onChange={(e) => updateTier(i, "rate_pct", Number(e.target.value) / 100)}
                    />
                  </td>
                  <td>
                    <button
                      type="button"
                      className="btn ghost small"
                      onClick={() => setTiers((ts) => ts.filter((_, idx) => idx !== i))}
                    >
                      ✕
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <button
            type="button"
            className="btn ghost small"
            onClick={() => setTiers((t) => [...t, blankTier()])}
          >
            + Add tier
          </button>

          <div className="form-actions">
            <button className="btn primary" type="submit">
              Create plan
            </button>
          </div>
        </form>
      )}

      <div className="plan-grid">
        {plans.map((p) => (
          <div className="plan-card" key={p.id}>
            <div className="plan-card-head">
              <h3>{p.name}</h3>
              <button className="btn ghost small" onClick={() => remove(p.id)}>
                Delete
              </button>
            </div>
            <div className="plan-meta">
              <span className="pill">{p.period_type}</span>
              <span>Quota {money(p.quota_amount)}</span>
              {p.base_salary > 0 && <span className="muted">Base {money(p.base_salary)}</span>}
            </div>
            <ul className="tier-list">
              {p.tiers.map((t, i) => (
                <li key={i}>
                  {(t.from_pct * 100).toFixed(0)}%
                  {t.to_pct > 0 ? `–${(t.to_pct * 100).toFixed(0)}%` : "+"} of quota @{" "}
                  <strong>{(t.rate_pct * 100).toFixed(1)}%</strong>
                </li>
              ))}
            </ul>
          </div>
        ))}
        {plans.length === 0 && <div className="muted">No comp plans yet.</div>}
      </div>
    </div>
  );
}

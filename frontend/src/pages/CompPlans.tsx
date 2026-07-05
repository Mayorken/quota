import { useEffect, useState } from "react";
import { api, CompPlan, Tier } from "../api";
import { money } from "../format";

const blankTier = (): Tier => ({ from_pct: 0, to_pct: 0, rate_pct: 0.05 });

export default function CompPlans() {
  const [plans, setPlans] = useState<CompPlan[]>([]);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);

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

  const periodLabels: Record<string, string> = {
    monthly: "Monthly",
    quarterly: "Quarterly",
    annual: "Annual",
  };

  return (
    <div>
      <div className="page-head">
        <div>
          <h2>Comp Plans</h2>
          <p className="muted">{plans.length} plan{plans.length !== 1 ? "s" : ""} configured</p>
        </div>
        <button className="btn primary" onClick={() => setShowForm((s) => !s)}>
          {showForm ? (
            "Cancel"
          ) : (
            <>
              <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14">
                <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
              </svg>
              New Plan
            </>
          )}
        </button>
      </div>

      {error && <div className="error">{error}</div>}

      {showForm && (
        <form className="card" onSubmit={submit}>
          <div className="grid-2">
            <div className="field">
              <label>Plan name</label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. AE Monthly Plan"
                required
              />
            </div>
            <div className="field">
              <label>Period</label>
              <select value={period} onChange={(e) => setPeriod(e.target.value as "monthly" | "quarterly" | "annual")}>
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
                placeholder="100,000"
                required
              />
            </div>
            <div className="field">
              <label>Base / draw ($)</label>
              <input
                type="number"
                value={base}
                onChange={(e) => setBase(e.target.value)}
                placeholder="Optional"
              />
            </div>
          </div>

          <h4 style={{ marginTop: 20, marginBottom: 4 }}>Commission Tiers</h4>
          <p className="muted small" style={{ marginBottom: 12 }}>
            Bands are expressed as % of quota. Set the top tier&apos;s &quot;To&quot; to 0 for uncapped accelerator.
          </p>

          <table className="tier-table">
            <thead>
              <tr>
                <th>From (% quota)</th>
                <th>To (% quota, 0 = uncapped)</th>
                <th>Rate (%)</th>
                <th style={{ width: 40 }}></th>
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
                    {tiers.length > 1 && (
                      <button
                        type="button"
                        className="btn ghost small"
                        onClick={() => setTiers((ts) => ts.filter((_, idx) => idx !== i))}
                        title="Remove tier"
                      >
                        <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14">
                          <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                        </svg>
                      </button>
                    )}
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
              Create Plan
            </button>
          </div>
        </form>
      )}

      <div className="plan-grid">
        {plans.map((p) => (
          <div className="plan-card" key={p.id}>
            <div className="plan-card-head">
              <h3>{p.name}</h3>
              <button className="btn ghost small danger" onClick={() => remove(p.id)}>
                Delete
              </button>
            </div>
            <div className="plan-meta">
              <span className="pill">{periodLabels[p.period_type] || p.period_type}</span>
              <span style={{ fontWeight: 600 }}>Quota {money(p.quota_amount)}</span>
              {p.base_salary > 0 && <span className="muted">Base {money(p.base_salary)}</span>}
            </div>
            <ul className="tier-list">
              {p.tiers.map((t, i) => (
                <li key={i}>
                  {(t.from_pct * 100).toFixed(0)}%
                  {t.to_pct > 0 ? ` \u2013 ${(t.to_pct * 100).toFixed(0)}%` : "+"} of quota @{" "}
                  <strong>{(t.rate_pct * 100).toFixed(1)}%</strong>
                </li>
              ))}
            </ul>
          </div>
        ))}
        {plans.length === 0 && (
          <div className="empty-card">
            No comp plans yet. Create one to define how commission is earned.
          </div>
        )}
      </div>
    </div>
  );
}

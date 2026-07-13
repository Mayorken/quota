import { useEffect, useState } from "react";
import {
  api,
  CommissionCalculation,
  CommissionStatus,
} from "../api";
import { useAuth, isManager } from "../auth";
import { money2, pct, shortDate } from "../format";

const STATUS_LABEL: Record<CommissionStatus, string> = {
  draft: "Draft",
  approved: "Approved",
  paid: "Paid",
};

// The next status a manager can advance a payout to, plus the button label.
const NEXT_ACTION: Partial<Record<CommissionStatus, { to: CommissionStatus; label: string }>> = {
  draft: { to: "approved", label: "Approve" },
  approved: { to: "paid", label: "Mark paid" },
};

export default function Payouts() {
  const { user } = useAuth();
  const manager = isManager(user);
  const [calcs, setCalcs] = useState<CommissionCalculation[]>([]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState<string | null>(null);
  const [date, setDate] = useState<string>("");
  const [filter, setFilter] = useState<CommissionStatus | "all">("all");

  function reload() {
    api
      .get<CommissionCalculation[]>("/commissions")
      .then(setCalcs)
      .catch((e) => setError(e.message));
  }

  useEffect(reload, []);

  async function generate() {
    setError("");
    setBusy("generate");
    try {
      const q = date ? `?date=${date}` : "";
      await api.post(`/commissions/generate${q}`);
      reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to generate");
    } finally {
      setBusy(null);
    }
  }

  async function transition(id: string, status: CommissionStatus) {
    setError("");
    setBusy(id);
    try {
      await api.post(`/commissions/${id}/transition`, { status });
      reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to update");
    } finally {
      setBusy(null);
    }
  }

  const rows = filter === "all" ? calcs : calcs.filter((c) => c.status === filter);
  const totals = rows.reduce(
    (acc, c) => {
      acc.owed += c.commission_owed;
      if (c.status === "approved") acc.approved += c.commission_owed;
      if (c.status === "paid") acc.paid += c.commission_owed;
      return acc;
    },
    { owed: 0, approved: 0, paid: 0 }
  );

  return (
    <div>
      <div className="page-head">
        <div>
          <h2>Payouts</h2>
          <p className="muted">
            {manager
              ? "Review, approve, and pay out commission for each period"
              : "Your finalized commission payouts"}
          </p>
        </div>
        {manager && (
          <div className="head-actions">
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              title="Snapshot a specific period"
            />
            <button className="btn primary" onClick={generate} disabled={busy === "generate"}>
              {busy === "generate" ? "Generating..." : "Generate period"}
            </button>
          </div>
        )}
      </div>

      {error && <div className="error">{error}</div>}

      {manager && rows.length > 0 && (
        <div className="stat-row">
          <div className="stat-card">
            <div className="stat-label">Total Commission</div>
            <div className="stat-value">{money2(totals.owed)}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Approved</div>
            <div className="stat-value">{money2(totals.approved)}</div>
          </div>
          <div className="stat-card highlight">
            <div className="stat-label">Paid</div>
            <div className="stat-value">{money2(totals.paid)}</div>
          </div>
        </div>
      )}

      <div className="head-actions" style={{ margin: "0 0 16px" }}>
        {(["all", "draft", "approved", "paid"] as const).map((f) => (
          <button
            key={f}
            className={`btn small ${filter === f ? "primary" : "ghost"}`}
            onClick={() => setFilter(f)}
          >
            {f === "all" ? "All" : STATUS_LABEL[f]}
          </button>
        ))}
      </div>

      <table className="data-table">
        <thead>
          <tr>
            {manager && <th>Rep</th>}
            <th>Period</th>
            <th className="num">Attainment</th>
            <th className="num">Commission</th>
            <th>Status</th>
            <th>Approved by</th>
            {manager && <th className="num" style={{ width: 200 }}></th>}
          </tr>
        </thead>
        <tbody>
          {rows.map((c) => {
            const next = NEXT_ACTION[c.status];
            const attainmentPct = c.quota ? (c.attained / c.quota) * 100 : 0;
            return (
              <tr key={c.id}>
                {manager && (
                  <td style={{ fontWeight: 500 }}>{c.rep?.name ?? c.rep_id}</td>
                )}
                <td>
                  {shortDate(c.period_start)} &ndash; {shortDate(c.period_end)}
                </td>
                <td className="num">{pct(attainmentPct)}</td>
                <td className="num" style={{ fontWeight: 600 }}>{money2(c.commission_owed)}</td>
                <td>
                  <span className={`status-badge status-${c.status}`}>
                    {STATUS_LABEL[c.status]}
                  </span>
                </td>
                <td className="muted">
                  {c.approved_by?.name ??
                    (c.approved_at ? shortDate(c.approved_at) : "\u2014")}
                </td>
                {manager && (
                  <td className="num">
                    {next && (
                      <button
                        className="btn small primary"
                        disabled={busy === c.id}
                        onClick={() => transition(c.id, next.to)}
                      >
                        {next.label}
                      </button>
                    )}
                    {c.status !== "draft" && c.status !== "paid" && (
                      <button
                        className="btn small ghost"
                        disabled={busy === c.id}
                        onClick={() => transition(c.id, "draft")}
                        style={{ marginLeft: 8 }}
                      >
                        Reopen
                      </button>
                    )}
                  </td>
                )}
              </tr>
            );
          })}
          {rows.length === 0 && (
            <tr>
              <td colSpan={manager ? 7 : 5} className="muted center" style={{ padding: "32px 16px" }}>
                {manager
                  ? "No payouts yet. Click \u201cGenerate period\u201d to snapshot this period\u2019s commission for the whole team."
                  : "No finalized payouts yet."}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, API_BASE, DashboardResponse, RepResult } from "../api";
import { useAuth, isManager } from "../auth";
import { money, pct, shortDate } from "../format";
import ProgressBar from "../components/ProgressBar";

export default function Dashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [data, setData] = useState<DashboardResponse | null>(null);
  const [error, setError] = useState("");
  const [date, setDate] = useState<string>("");

  useEffect(() => {
    const q = date ? `?date=${date}` : "";
    api
      .get<DashboardResponse>(`/dashboard${q}`)
      .then(setData)
      .catch((e) => setError(e.message));
  }, [date]);

  if (error) return <div className="error">{error}</div>;
  if (!data) return <div className="muted">Loading dashboard…</div>;

  const results = [...data.results].sort((a, b) => b.attainment_pct - a.attainment_pct);
  const withPlan = results.filter((r) => r.has_plan);
  const teamQuota = withPlan.reduce((s, r) => s + r.quota, 0);
  const teamAttained = withPlan.reduce((s, r) => s + r.attained, 0);
  const teamCommission = withPlan.reduce((s, r) => s + r.commission_owed, 0);

  function downloadCSV() {
    const token = localStorage.getItem("quota_token");
    const q = date ? `?date=${date}` : "";
    fetch(`${API_BASE}/api/export/commissions.csv${q}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.blob())
      .then((blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = "commissions.csv";
        a.click();
        URL.revokeObjectURL(url);
      });
  }

  return (
    <div>
      <div className="page-head">
        <div>
          <h2>{isManager(user) ? "Team dashboard" : "My attainment"}</h2>
          <p className="muted">
            Period as of {shortDate(data.as_of)}
            {withPlan[0] && ` · ${shortDate(withPlan[0].period_start)} – ${shortDate(withPlan[0].period_end)}`}
          </p>
        </div>
        <div className="head-actions">
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            title="View a past period"
          />
          {isManager(user) && (
            <button className="btn primary" onClick={downloadCSV}>
              Export payroll CSV
            </button>
          )}
        </div>
      </div>

      {isManager(user) && withPlan.length > 0 && (
        <div className="stat-row">
          <div className="stat-card">
            <div className="stat-label">Team quota</div>
            <div className="stat-value">{money(teamQuota)}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Team attained</div>
            <div className="stat-value">{money(teamAttained)}</div>
            <div className="stat-sub">{pct(teamQuota ? (teamAttained / teamQuota) * 100 : 0)}</div>
          </div>
          <div className="stat-card highlight">
            <div className="stat-label">Commission owed</div>
            <div className="stat-value">{money(teamCommission)}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Reps on plan</div>
            <div className="stat-value">{withPlan.length}</div>
          </div>
        </div>
      )}

      <div className="leaderboard">
        {results.map((r, i) => (
          <RepRow key={r.rep_id} rank={i + 1} r={r} onClick={() => navigate(`/reps/${r.rep_id}`)} />
        ))}
        {results.length === 0 && <div className="muted">No reps yet.</div>}
      </div>
    </div>
  );
}

function RepRow({ rank, r, onClick }: { rank: number; r: RepResult; onClick: () => void }) {
  if (!r.has_plan) {
    return (
      <div className="rep-row disabled">
        <div className="rank">—</div>
        <div className="rep-name">{r.rep_name}</div>
        <div className="muted">No active comp plan</div>
      </div>
    );
  }
  return (
    <div className="rep-row clickable" onClick={onClick}>
      <div className="rank">#{rank}</div>
      <div className="rep-info">
        <div className="rep-name">{r.rep_name}</div>
        <div className="rep-plan muted">{r.comp_plan_name}</div>
      </div>
      <div className="rep-bar">
        <ProgressBar pct={r.attainment_pct} />
        <div className="bar-labels">
          <span>{money(r.attained)}</span>
          <span className="muted">quota {money(r.quota)}</span>
        </div>
      </div>
      <div className="rep-attain">
        <div className="attain-pct">{pct(r.attainment_pct)}</div>
        <div className="muted">attainment</div>
      </div>
      <div className="rep-comm">
        <div className="comm-value">{money(r.commission_owed)}</div>
        <div className="muted">commission</div>
      </div>
    </div>
  );
}

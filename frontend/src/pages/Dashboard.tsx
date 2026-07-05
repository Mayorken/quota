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
  if (!data) return <div className="muted" style={{ padding: "40px 0" }}>Loading dashboard...</div>;

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
          <h2>{isManager(user) ? "Team Dashboard" : "My Attainment"}</h2>
          <p className="muted">
            {withPlan[0]
              ? `${shortDate(withPlan[0].period_start)} - ${shortDate(withPlan[0].period_end)}`
              : `As of ${shortDate(data.as_of)}`}
          </p>
        </div>
        <div className="head-actions">
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            title="View a different period"
          />
          {isManager(user) && (
            <button className="btn primary" onClick={downloadCSV}>
              <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14">
                <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
              Export CSV
            </button>
          )}
        </div>
      </div>

      {isManager(user) && withPlan.length > 0 && (
        <div className="stat-row">
          <div className="stat-card">
            <div className="stat-label">Team Quota</div>
            <div className="stat-value">{money(teamQuota)}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Attained</div>
            <div className="stat-value">{money(teamAttained)}</div>
            <div className="stat-sub">{pct(teamQuota ? (teamAttained / teamQuota) * 100 : 0)} of target</div>
          </div>
          <div className="stat-card highlight">
            <div className="stat-label">Commission Owed</div>
            <div className="stat-value">{money(teamCommission)}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Active Reps</div>
            <div className="stat-value">{withPlan.length}</div>
            <div className="stat-sub">{results.length - withPlan.length > 0 ? `${results.length - withPlan.length} unassigned` : "All on plan"}</div>
          </div>
        </div>
      )}

      <div className="leaderboard">
        {results.map((r, i) => (
          <RepRow key={r.rep_id} rank={i + 1} r={r} onClick={() => navigate(`/reps/${r.rep_id}`)} />
        ))}
        {results.length === 0 && (
          <div className="empty-card">
            No reps yet. Add team members and assign comp plans to get started.
          </div>
        )}
      </div>
    </div>
  );
}

function RepRow({ rank, r, onClick }: { rank: number; r: RepResult; onClick: () => void }) {
  if (!r.has_plan) {
    return (
      <div className="rep-row disabled">
        <div className="rank">&mdash;</div>
        <div className="rep-info">
          <div className="rep-name">{r.rep_name}</div>
          <div className="rep-plan muted">No active comp plan</div>
        </div>
        <div />
        <div />
        <div />
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
          <span>{money(r.quota)} quota</span>
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

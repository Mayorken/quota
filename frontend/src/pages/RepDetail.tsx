import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { api, RepResult } from "../api";
import { money, money2, pct, shortDate } from "../format";
import ProgressBar from "../components/ProgressBar";

export default function RepDetail() {
  const { id } = useParams();
  const [r, setR] = useState<RepResult | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .get<RepResult>(`/reps/${id}/commission`)
      .then(setR)
      .catch((e) => setError(e.message));
  }, [id]);

  if (error) return <div className="error">{error}</div>;
  if (!r) return <div className="muted">Loading…</div>;

  return (
    <div>
      <Link to="/" className="back-link">
        ← Back to dashboard
      </Link>
      <div className="page-head">
        <div>
          <h2>{r.rep_name}</h2>
          <p className="muted">{r.rep_email}</p>
        </div>
      </div>

      {!r.has_plan ? (
        <div className="empty-card">This rep has no active comp plan.</div>
      ) : (
        <>
          <div className="detail-head-card">
            <div className="dh-item">
              <div className="muted">Plan</div>
              <div className="dh-value">{r.comp_plan_name}</div>
            </div>
            <div className="dh-item">
              <div className="muted">Period</div>
              <div className="dh-value">
                {shortDate(r.period_start)} – {shortDate(r.period_end)}
              </div>
            </div>
            <div className="dh-item">
              <div className="muted">Attainment</div>
              <div className="dh-value">{pct(r.attainment_pct)}</div>
            </div>
            <div className="dh-item highlight">
              <div className="muted">Commission owed</div>
              <div className="dh-value">{money2(r.commission_owed)}</div>
            </div>
          </div>

          <div className="bar-block">
            <ProgressBar pct={r.attainment_pct} />
            <div className="bar-labels">
              <span>{money(r.attained)} attained</span>
              <span className="muted">quota {money(r.quota)}</span>
            </div>
          </div>

          <h3>How this was calculated</h3>
          <p className="muted">
            Every rep and manager sees this exact math — no hidden spreadsheet formulas.
          </p>

          <div className="calc-summary">
            <Row label="Total revenue closed" value={money2(r.breakdown.total_revenue)} />
            {r.breakdown.credited_revenue !== r.breakdown.total_revenue && (
              <Row
                label="Credited revenue (after deal-type multipliers)"
                value={money2(r.breakdown.credited_revenue)}
              />
            )}
            <Row label="Deals closed" value={String(r.breakdown.deal_count)} />
            <Row label="Quota" value={money2(r.breakdown.quota)} />
          </div>

          <table className="calc-table">
            <thead>
              <tr>
                <th>Tier</th>
                <th className="num">Revenue in band</th>
                <th className="num">Rate</th>
                <th className="num">Commission</th>
              </tr>
            </thead>
            <tbody>
              {r.breakdown.lines.map((line, i) => (
                <tr key={i} className={line.revenue_in_band > 0 ? "" : "dim"}>
                  <td>{line.label}</td>
                  <td className="num">{money2(line.revenue_in_band)}</td>
                  <td className="num">{(line.rate_pct * 100).toFixed(1)}%</td>
                  <td className="num">{money2(line.commission)}</td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr>
                <td colSpan={3}>Total commission</td>
                <td className="num total">{money2(r.breakdown.total_commission)}</td>
              </tr>
            </tfoot>
          </table>
        </>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="calc-row">
      <span className="muted">{label}</span>
      <span className="mono">{value}</span>
    </div>
  );
}

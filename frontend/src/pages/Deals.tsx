import { useEffect, useState } from "react";
import { api, Deal, User } from "../api";
import { money, shortDate } from "../format";

export default function Deals() {
  const [deals, setDeals] = useState<Deal[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);

  const [repId, setRepId] = useState("");
  const [amount, setAmount] = useState("");
  const [dealType, setDealType] = useState("new");
  const [closeDate, setCloseDate] = useState(new Date().toISOString().slice(0, 10));

  function reload() {
    api.get<Deal[]>("/deals").then(setDeals).catch((e) => setError(e.message));
  }

  useEffect(() => {
    reload();
    api
      .get<User[]>("/users")
      .then((u) => {
        setUsers(u);
        const firstRep = u.find((x) => x.role === "rep");
        if (firstRep) setRepId(firstRep.id);
      })
      .catch((e) => setError(e.message));
  }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    try {
      await api.post("/deals", {
        rep_id: repId,
        amount: parseFloat(amount),
        deal_type: dealType,
        close_date: new Date(closeDate + "T12:00:00Z").toISOString(),
      });
      setAmount("");
      setShowForm(false);
      reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed");
    }
  }

  async function remove(id: string) {
    if (!confirm("Delete this deal?")) return;
    await api.del(`/deals/${id}`);
    reload();
  }

  return (
    <div>
      <div className="page-head">
        <div>
          <h2>Deals</h2>
          <p className="muted">{deals.length} deal{deals.length !== 1 ? "s" : ""} recorded</p>
        </div>
        <button className="btn primary" onClick={() => setShowForm((s) => !s)}>
          {showForm ? (
            "Cancel"
          ) : (
            <>
              <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14">
                <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
              </svg>
              Add Deal
            </>
          )}
        </button>
      </div>

      {error && <div className="error">{error}</div>}

      {showForm && (
        <form className="card form-inline" onSubmit={submit}>
          <div className="field">
            <label>Rep</label>
            <select value={repId} onChange={(e) => setRepId(e.target.value)} required>
              {users
                .filter((u) => u.role === "rep")
                .map((u) => (
                  <option key={u.id} value={u.id}>
                    {u.name}
                  </option>
                ))}
            </select>
          </div>
          <div className="field">
            <label>Amount ($)</label>
            <input
              type="number"
              step="0.01"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="50,000"
              required
            />
          </div>
          <div className="field">
            <label>Deal type</label>
            <select value={dealType} onChange={(e) => setDealType(e.target.value)}>
              <option value="new">New business</option>
              <option value="expansion">Expansion</option>
              <option value="renewal">Renewal</option>
            </select>
          </div>
          <div className="field">
            <label>Close date</label>
            <input
              type="date"
              value={closeDate}
              onChange={(e) => setCloseDate(e.target.value)}
              required
            />
          </div>
          <button className="btn primary" type="submit">
            Save Deal
          </button>
        </form>
      )}

      <table className="data-table">
        <thead>
          <tr>
            <th>Rep</th>
            <th className="num">Amount</th>
            <th>Type</th>
            <th>Close Date</th>
            <th className="num" style={{ width: 80 }}></th>
          </tr>
        </thead>
        <tbody>
          {deals.map((d) => (
            <tr key={d.id}>
              <td style={{ fontWeight: 500 }}>{d.rep?.name ?? d.rep_id}</td>
              <td className="num" style={{ fontWeight: 600 }}>{money(d.amount)}</td>
              <td>
                <span className="pill">{d.deal_type || "\u2014"}</span>
              </td>
              <td>{shortDate(d.close_date)}</td>
              <td className="num">
                <button className="btn ghost small danger" onClick={() => remove(d.id)}>
                  Delete
                </button>
              </td>
            </tr>
          ))}
          {deals.length === 0 && (
            <tr>
              <td colSpan={5} className="muted center" style={{ padding: "32px 16px" }}>
                No deals recorded yet. Add one to see attainment update in real time.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

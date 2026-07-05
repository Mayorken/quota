import { useEffect, useState } from "react";
import { api, Deal, User } from "../api";
import { money, shortDate } from "../format";

export default function Deals() {
  const [deals, setDeals] = useState<Deal[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState("");
  const [showForm, setShowForm] = useState(false);

  // Form state.
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
        <h2>Deals</h2>
        <button className="btn primary" onClick={() => setShowForm((s) => !s)}>
          {showForm ? "Cancel" : "+ Add deal"}
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
              required
            />
          </div>
          <div className="field">
            <label>Deal type</label>
            <select value={dealType} onChange={(e) => setDealType(e.target.value)}>
              <option value="new">new</option>
              <option value="expansion">expansion</option>
              <option value="renewal">renewal</option>
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
            Save
          </button>
        </form>
      )}

      <table className="data-table">
        <thead>
          <tr>
            <th>Rep</th>
            <th className="num">Amount</th>
            <th>Type</th>
            <th>Close date</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {deals.map((d) => (
            <tr key={d.id}>
              <td>{d.rep?.name ?? d.rep_id}</td>
              <td className="num">{money(d.amount)}</td>
              <td>
                <span className="pill">{d.deal_type || "—"}</span>
              </td>
              <td>{shortDate(d.close_date)}</td>
              <td className="num">
                <button className="btn ghost small" onClick={() => remove(d.id)}>
                  Delete
                </button>
              </td>
            </tr>
          ))}
          {deals.length === 0 && (
            <tr>
              <td colSpan={5} className="muted center">
                No deals yet. Add one to see attainment update.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}

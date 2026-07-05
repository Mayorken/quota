export default function ProgressBar({ pct }: { pct: number }) {
  const base = Math.min(pct, 100);
  const over = Math.max(0, Math.min(pct - 100, 100));

  const color =
    pct >= 100
      ? "var(--green)"
      : pct >= 70
        ? "var(--blue)"
        : pct >= 40
          ? "var(--amber)"
          : "var(--red)";

  return (
    <div className="progress-track">
      <div
        className="progress-fill"
        style={{ width: `${base}%`, background: color }}
      />
      {over > 0 && (
        <div
          className="progress-over"
          style={{ width: `${over / 2}%` }}
          title={`${pct.toFixed(0)}% of quota`}
        />
      )}
    </div>
  );
}

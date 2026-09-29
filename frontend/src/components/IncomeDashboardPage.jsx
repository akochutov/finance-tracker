import { useState, useEffect } from "react";
import { getIncomeDashboard } from "../api/client";
import { formatWhole, formatCompact, niceMax } from "./dashboardUtils";

const CHART_HEIGHT = 220;

function monthName(key) {
    const [y, m] = key.split("-").map(Number);
    return new Date(y, m - 1, 1).toLocaleDateString(undefined, { month: "short", year: "numeric" });
}

function BarList({ title, meta, items, total, color }) {
    const max = items.length > 0 ? Number(items[0].total) : 1;
    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>{title}</h5>
                <span className="row-meta">{meta}</span>
            </div>
            <div className="id-list">
                {items.map((it) => (
                    <div key={it.key} className="id-bar-row">
                        <span title={it.name}>{it.name}</span>
                        <span className="xd-track">
                            <span style={{ width: `${Math.max((Number(it.total) / max) * 100, 1)}%`, background: color }} />
                        </span>
                        <span className="num mono">{formatWhole(it.total)}</span>
                        <span className="num row-meta">
                            {total > 0 ? ((Number(it.total) / total) * 100).toFixed(1) : "0.0"}%
                        </span>
                    </div>
                ))}
                {items.length === 0 && <p className="xd-note">No income yet.</p>}
            </div>
        </section>
    );
}

function Metric({ label, value, sub }) {
    return (
        <div className="xd-panel">
            <div className="xd-kpi-label">{label}</div>
            <div className="xd-kpi-value mono">{formatWhole(value)}</div>
            <div className="xd-kpi-sub">{sub}</div>
        </div>
    );
}

function IncomeDashboardPage() {
    const [data, setData] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        getIncomeDashboard()
            .then(setData)
            .catch((err) => setError(err.message));
    }, []);

    const header = (
        <div className="page-header">
            <h1>Income dashboard</h1>
            <p>Income, converted{data ? ` to ${data.currency}` : ""} at each payment's date.</p>
        </div>
    );

    if (error) {
        return <div className="xd-page">{header}<div className="error">{error}</div></div>;
    }
    if (!data) {
        return <div className="xd-page">{header}<div className="loading">Loading…</div></div>;
    }

    const total = Number(data.total_income);
    const now = new Date();
    const currentYear = String(now.getFullYear());
    const monthsThisYear = now.getMonth() + 1;
    const lastMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
    const fmtMonth = (d) => d.toLocaleDateString(undefined, { month: "long", year: "numeric" });

    const max = niceMax(Math.max(0, ...data.by_month.map((m) => Number(m.total))));
    const currencies = data.by_currency.map((c) => ({ key: c.currency, name: c.currency, total: c.total }));
    const companies = data.by_company.map((c) => ({
        key: c.company_id,
        name: c.name || c.company_id.slice(0, 8),
        total: c.total,
    }));

    return (
        <div className="xd-page">
            {header}

            <section className="xd-panel id-total">
                <div>
                    <div className="xd-kpi-label">Total income · all time</div>
                    <div className="id-total-value mono">{formatWhole(total)}</div>
                </div>
                <div className="id-stats">
                    <div>
                        <div className="xd-kpi-label">Records</div>
                        <div className="id-stat-value">{data.records.total}</div>
                    </div>
                    <div>
                        <div className="xd-kpi-label">Not converted</div>
                        <div className={data.records.unconverted > 0 ? "id-stat-value xd-up" : "id-stat-value"}>
                            {data.records.unconverted}
                        </div>
                    </div>
                    {data.period.from && (
                        <div>
                            <div className="xd-kpi-label">Period</div>
                            <div className="id-stat-value">
                                {monthName(data.period.from)} – {monthName(data.period.to)}
                            </div>
                        </div>
                    )}
                </div>
            </section>

            <section className="xd-panel">
                <div className="xd-panel-head">
                    <h5>By month · all time</h5>
                    <div className="xd-legend">
                        <span><i style={{ background: "var(--color-accent)" }} />{currentYear}</span>
                        <span><i style={{ background: "var(--color-accent-300)" }} />Earlier years</span>
                    </div>
                </div>
                <div className="id-chart">
                    <div className="id-axis">
                        <span>{formatCompact(max)}</span>
                        <span>{formatCompact(max / 2)}</span>
                        <span>0</span>
                    </div>
                    <div className="id-plot">
                        <div className="id-bars">
                            {data.by_month.map((m) => (
                                <div
                                    key={m.month}
                                    className="id-bar"
                                    title={`${monthName(m.month)}: ${formatWhole(m.total)}`}
                                    style={{
                                        height: `${(Number(m.total) / max) * CHART_HEIGHT}px`,
                                        background: m.month.startsWith(currentYear)
                                            ? "var(--color-accent)"
                                            : "var(--color-accent-300)",
                                    }}
                                />
                            ))}
                        </div>
                        <div className="id-years">
                            {data.by_month.map((m, i) => (
                                <div key={m.month} className="id-year">
                                    {i === 0 || m.month.endsWith("-01") ? m.month.slice(0, 4) : ""}
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </section>

            <div className="xd-row-halves">
                <BarList
                    title={`By currency · in ${data.currency}`}
                    meta={`${currencies.length} currencies`}
                    items={currencies}
                    total={total}
                    color="var(--color-accent-700)"
                />
                <BarList
                    title={`By company · in ${data.currency}`}
                    meta={`${companies.length} payers`}
                    items={companies}
                    total={total}
                    color="#e8590c"
                />
            </div>

            <div className="id-metrics">
                <Metric
                    label="Avg / month · all time"
                    value={data.averages.monthly_all_time}
                    sub={`over ${data.by_month.length} months`}
                />
                <Metric
                    label="Avg / month · this year"
                    value={data.averages.monthly_this_year}
                    sub={`over ${monthsThisYear} months`}
                />
                <Metric label="Total · this year" value={data.periods.this_year} sub={currentYear} />
                <Metric label="Last month" value={data.periods.last_month} sub={fmtMonth(lastMonth)} />
                <Metric label="This month" value={data.periods.this_month} sub={fmtMonth(now)} />
            </div>

            {data.records.unconverted > 0 && (
                <p className="xd-note">
                    {data.records.unconverted} incomes have no exchange rate for their date and are left out of the totals.
                </p>
            )}
        </div>
    );
}

export default IncomeDashboardPage;
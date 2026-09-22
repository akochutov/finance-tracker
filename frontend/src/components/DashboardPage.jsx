import { useState, useEffect } from "react";
import { getDashboard } from "../api/client";

function money(value, currency, opts = {}) {
    const n = Number(value);
    try {
        return new Intl.NumberFormat("en-US", {
            style: "currency", currency, maximumFractionDigits: 0, ...opts,
        }).format(n);
    } catch {
        return `${Math.round(n).toLocaleString("en-US")} ${currency}`;
    }
}

function compactMoney(value, currency) {
    return money(value, currency, { notation: "compact", maximumFractionDigits: 1 });
}

function fmtMonth(key) {
    if (!key) return "";
    const [y, m] = key.split("-");
    return new Date(Number(y), Number(m) - 1, 1)
        .toLocaleString("en-US", { month: "short", year: "numeric" });
}

function MetricCard({ label, value, currency }) {
    return (
        <div className="metric-card">
            <div className="metric-label">{label}</div>
            <div className="metric-value">{money(value, currency)}</div>
        </div>
    );
}

function CategoryBars({ title, items, currency }) {
    const max = Math.max(1, ...items.map((it) => Number(it.value)));
    return (
        <div className="card">
            <div className="chart-title">{title}</div>
            <div className="bars">
                {items.map((it, i) => (
                    <div
                        key={i}
                        className="bar"
                        style={{ height: `${(Number(it.value) / max) * 100}%` }}
                        title={`${it.label}: ${money(it.value, currency)}`}
                    />
                ))}
            </div>
            <div className="bar-labels">
                {items.map((it, i) => (
                    <div key={i} className="bl">
                        {it.label}<br /><span className="blv">{compactMoney(it.value, currency)}</span>
                    </div>
                ))}
            </div>
        </div>
    );
}

function MonthlyChart({ data, currency }) {
    const max = Math.max(1, ...data.map((d) => Number(d.total)));
    const years = [...new Set(data.map((d) => d.month.slice(0, 4)))];
    return (
        <div className="card" style={{ marginBottom: 12 }}>
            <div className="chart-title">By month · all time ({currency})</div>
            <div className="month-bars">
                {data.map((d) => (
                    <div
                        key={d.month}
                        className="bar"
                        style={{ height: `${(Number(d.total) / max) * 100}%` }}
                        title={`${fmtMonth(d.month)}: ${money(d.total, currency)}`}
                    />
                ))}
            </div>
            <div className="month-axis">
                {years.map((y) => <span key={y}>{y}</span>)}
            </div>
        </div>
    );
}

function DashboardPage() {
    const [data, setData] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        getDashboard().then(setData).catch((err) => setError(err.message));
    }, []);

    if (error) {
        return (<><div className="page-header"><h1>Dashboard</h1></div><div className="error">{error}</div></>);
    }
    if (!data) {
        return (<><div className="page-header"><h1>Dashboard</h1></div><div className="loading">Loading...</div></>);
    }

    const cur = data.currency;
    const rec = data.records;

    if (rec.total === 0) {
        return (
            <>
                <div className="page-header"><h1>Dashboard</h1><p>Income converted to {cur}.</p></div>
                <div className="loading">No income data yet.</div>
            </>
        );
    }

    const currencyItems = data.by_currency.map((c) => ({ label: c.currency, value: c.total }));
    const companyItems = data.by_company.map((c) => ({
        label: c.name || c.company_id.slice(0, 6),
        value: c.total,
    }));

    return (
        <>
            <div className="page-header">
                <h1>Dashboard</h1>
                <p>Income converted to {cur}.</p>
            </div>

            <div className="dash-row-2">
                <div className="card dash-total-card">
                    <div className="dash-total-label">Total income · all time</div>
                    <div className="dash-total-value">{money(data.total_income, cur)}</div>
                    <div className="dash-total-sub">
                        {cur} · {fmtMonth(data.period.from)} – {fmtMonth(data.period.to)}
                    </div>
                </div>

                <div className="card dash-records">
                    <div className="rec-row"><span className="k">Incomes</span><span className="v">{rec.total}</span></div>
                    <div className="rec-row"><span className="k">Converted</span><span className="v">{rec.converted}</span></div>
                    <div className="rec-row"><span className="k">Not converted</span><span className="v">{rec.unconverted}</span></div>
                    <div className="rec-row">
                        <span className="k">Companies · currencies</span>
                        <span className="v">{data.by_company.length} · {data.by_currency.length}</span>
                    </div>
                </div>
            </div>

            <MonthlyChart data={data.by_month} currency={cur} />

            <div className="dash-row-2">
                <CategoryBars title={`By currency · in ${cur}`} items={currencyItems} currency={cur} />
                <CategoryBars title={`By company · in ${cur}`} items={companyItems} currency={cur} />
            </div>

            <div className="dash-row-5">
                <MetricCard label="Avg / month · all time" value={data.averages.monthly_all_time} currency={cur} />
                <MetricCard label="Avg / month · this year" value={data.averages.monthly_this_year} currency={cur} />
                <MetricCard label="Total · this year" value={data.periods.this_year} currency={cur} />
                <MetricCard label="Last month" value={data.periods.last_month} currency={cur} />
                <MetricCard label="This month" value={data.periods.this_month} currency={cur} />
            </div>
        </>
    );
}

export default DashboardPage;
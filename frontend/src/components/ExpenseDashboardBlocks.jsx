import {
    formatWhole, formatCompact, percentChange, niceMax, monthShort, OTHER_COLOR,
} from "./dashboardUtils";
import { formatISODate } from "./expenseUtils";

const CHART_HEIGHT = 220;

export function Delta({ current, previous }) {
    const pct = percentChange(current, previous);
    if (pct === null) {
        return <span className="xd-delta">new</span>;
    }
    const up = pct >= 0;
    return (
        <span className={up ? "xd-delta xd-up" : "xd-delta xd-down"}>
            {up ? "▲" : "▼"} {Math.abs(pct).toFixed(1)}%
        </span>
    );
}

export function KpiRow({ summary, compare, prevLabel }) {
    const total = Number(summary.total);
    const fixedShare = total > 0 ? (Number(summary.fixed) / total) * 100 : 0;
    const rate = summary.savings_rate;
    const prevRate = summary.previous_savings_rate;
    const income = Number(summary.income);

    return (
        <div className="xd-kpis">
            <div className="xd-panel">
                <div className="xd-kpi-label">Total spent</div>
                <div className="xd-kpi-value mono">{formatWhole(summary.total)}</div>
                {compare && (
                    <div className="xd-kpi-sub">
                        <Delta current={summary.total} previous={summary.previous_total} /> vs {prevLabel} · {formatWhole(summary.previous_total)}
                    </div>
                )}
            </div>

            <div className="xd-panel">
                <div className="xd-kpi-label">Average per day</div>
                <div className="xd-kpi-value mono">{formatWhole(summary.daily_average)}</div>
                <div className="xd-kpi-sub">
                    {compare && `${prevLabel} · ${formatWhole(summary.previous_daily_average)} · `}
                    {summary.receipts} receipts
                </div>
            </div>

            <div className="xd-panel">
                <div className="xd-kpi-label">Fixed costs</div>
                <div className="xd-kpi-value mono">{formatWhole(summary.fixed)}</div>
                <div className="xd-split"><div style={{ width: `${fixedShare}%` }} /></div>
                <div className="xd-kpi-sub">
                    {fixedShare.toFixed(1)}% fixed · {formatWhole(summary.variable)} variable
                </div>
            </div>

            <div className="xd-panel">
                <div className="xd-kpi-label">Savings rate</div>
                <div className="xd-kpi-value mono">{rate === null ? "—" : `${Number(rate).toFixed(1)}%`}</div>
                <div className="xd-kpi-sub">
                    {income > 0
                        ? `Left ${formatWhole(income - total)} of ${formatWhole(income)} income`
                        : "No income in this period"}
                    {compare && prevRate !== null && ` · ${prevLabel} ${Number(prevRate).toFixed(1)}%`}
                </div>
            </div>
        </div>
    );
}

export function GroupsBreakdown({ groups, total, selectedId, onSelect, colorOf, compare, prevLabel }) {
    const active = groups.filter((g) => Number(g.total) > 0);
    const gone = groups.filter((g) => Number(g.total) === 0);
    const max = active.length > 0 ? Number(active[0].total) : 1;

    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>By group</h5>
                <span className="row-meta">Click a group for its categories</span>
            </div>
            <div className={compare ? "xd-group-row xd-group-head" : "xd-group-row xd-group-head xd-no-compare"}>
                <span>Group</span>
                <span>Share</span>
                <span className="num">Amount</span>
                <span className="num">%</span>
                {compare && <span className="num">vs {prevLabel}</span>}
            </div>
            {active.map((g) => {
                const diff = Number(g.total) - Number(g.previous);
                const rowClass = [
                    "xd-group-row",
                    compare ? "" : "xd-no-compare",
                    g.group_id === selectedId ? "selected" : "",
                ].join(" ");
                return (
                    <button type="button" key={g.group_id} className={rowClass} onClick={() => onSelect(g.group_id)}>
                        <span className="xd-group-name">{g.name}</span>
                        <span className="xd-track">
                            <span style={{ width: `${Math.max((Number(g.total) / max) * 100, 0.6)}%`, background: colorOf(g.group_id) }} />
                        </span>
                        <span className="num mono">{formatWhole(g.total)}</span>
                        <span className="num row-meta">{((Number(g.total) / total) * 100).toFixed(1)}%</span>
                        {compare && (
                            <span className={diff >= 0 ? "num mono xd-up" : "num mono xd-down"}>
                                {diff >= 0 ? "+" : "−"}{formatWhole(Math.abs(diff))}
                            </span>
                        )}
                    </button>
                );
            })}
            {compare && gone.length > 0 && (
                <p className="xd-note">
                    Not in this period: {gone.map((g) => `${g.name} (${prevLabel} ${formatWhole(g.previous)})`).join(", ")}.
                </p>
            )}
        </section>
    );
}

export function CategoriesPanel({ group, color }) {
    if (!group) {
        return <section className="xd-panel"><p className="xd-note">No spending in this period.</p></section>;
    }
    const cats = group.categories;
    const shown = cats.slice(0, 10);
    const rest = cats.slice(10);
    const max = shown.length > 0 ? Number(shown[0].total) : 1;

    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>Categories</h5>
                <span className="row-meta">{cats.length} {cats.length === 1 ? "category" : "categories"}</span>
            </div>
            <div className="xd-cat-title">
                <span>{group.name}</span>
                <span className="mono">{formatWhole(group.total)}</span>
            </div>
            <div className="xd-cats">
                {shown.map((c) => (
                    <div key={c.category_id} className="xd-cat">
                        <div className="xd-cat-line">
                            <span>{c.name}</span>
                            <span className="mono">
                                {formatWhole(c.total)}{" "}
                                <span className="row-meta">· {((Number(c.total) / Number(group.total)) * 100).toFixed(1)}%</span>
                            </span>
                        </div>
                        <div className="xd-track xd-track-thin">
                            <span style={{ width: `${Math.max((Number(c.total) / max) * 100, 1)}%`, background: color }} />
                        </div>
                    </div>
                ))}
            </div>
            {rest.length > 0 && (
                <p className="xd-note">
                    + {rest.length} more · {formatWhole(rest.reduce((s, c) => s + Number(c.total), 0))}
                </p>
            )}
        </section>
    );
}

function MonthChart({ max, children }) {
    return (
        <div className="xd-chart">
            <div className="xd-axis">
                <span>{formatCompact(max)}</span>
                <span>{formatCompact(max / 2)}</span>
                <span>0</span>
            </div>
            <div className="xd-cols">{children}</div>
        </div>
    );
}

const px = (value, max) => `${(Number(value) / max) * CHART_HEIGHT}px`;

export function TrendChart({ trend, topIds, colorOf, groupNames, inPeriod }) {
    const max = niceMax(Math.max(...trend.map((m) => Number(m.total))));
    const topSet = new Set(topIds);

    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>Monthly trend by group</h5>
                <div className="xd-legend">
                    {topIds.map((id) => (
                        <span key={id}><i style={{ background: colorOf(id) }} />{groupNames[id] || "…"}</span>
                    ))}
                    <span><i style={{ background: OTHER_COLOR }} />Other groups</span>
                </div>
            </div>
            <MonthChart max={max}>
                {trend.map((m, i) => {
                    const other = m.by_group
                        .filter((g) => !topSet.has(g.group_id))
                        .reduce((s, g) => s + Number(g.total), 0);
                    return (
                        <div key={m.month} className={inPeriod(m.month) ? "xd-col" : "xd-col muted"}>
                            <span className="xd-col-value mono">{Number(m.total) > 0 ? formatCompact(m.total) : ""}</span>
                            <div className="xd-stack">
                                {topIds.map((id) => {
                                    const g = m.by_group.find((x) => x.group_id === id);
                                    return g ? <div key={id} style={{ height: px(g.total, max), background: colorOf(id) }} /> : null;
                                })}
                                <div style={{ height: px(other, max), background: OTHER_COLOR }} />
                            </div>
                            <span className="xd-col-label">{monthShort(m.month, i === 0)}</span>
                        </div>
                    );
                })}
            </MonthChart>
        </section>
    );
}

export function FixedVariableChart({ trend, inPeriod }) {
    const max = niceMax(Math.max(...trend.map((m) => Number(m.total))));
    const withFixed = trend.filter((m) => Number(m.fixed) > 0).map((m) => Number(m.fixed));

    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>Fixed vs variable</h5>
                <div className="xd-legend">
                    <span><i style={{ background: "#2f49b0" }} />Fixed (lines with a period)</span>
                    <span><i style={{ background: "#a9b8f5" }} />Variable</span>
                </div>
            </div>
            <MonthChart max={max}>
                {trend.map((m, i) => (
                    <div key={m.month} className={inPeriod(m.month) ? "xd-col" : "xd-col muted"}>
                        <span className="xd-col-value mono xd-fixed-value">{Number(m.fixed) > 0 ? formatCompact(m.fixed) : ""}</span>
                        <div className="xd-stack xd-stack-narrow">
                            <div style={{ height: px(m.fixed, max), background: "#2f49b0" }} />
                            <div style={{ height: px(m.variable, max), background: "#a9b8f5" }} />
                        </div>
                        <span className="xd-col-label">{monthShort(m.month, i === 0)}</span>
                    </div>
                ))}
            </MonthChart>
            {withFixed.length > 0 && (
                <p className="xd-note">
                    Fixed base: {formatWhole(Math.min(...withFixed))} – {formatWhole(Math.max(...withFixed))} a month.
                </p>
            )}
        </section>
    );
}

export function CashflowChart({ trend, inPeriod }) {
    const max = niceMax(Math.max(...trend.map((m) => Math.max(Number(m.total), Number(m.income)))));
    const net = trend.reduce((s, m) => s + Number(m.income) - Number(m.total), 0);

    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>Cashflow</h5>
                <div className="xd-legend">
                    <span><i style={{ background: "#2f49b0" }} />Income</span>
                    <span><i style={{ background: "#e8590c" }} />Expenses</span>
                </div>
            </div>
            <MonthChart max={max}>
                {trend.map((m, i) => {
                    const diff = Number(m.income) - Number(m.total);
                    const empty = Number(m.income) === 0 && Number(m.total) === 0;
                    return (
                        <div key={m.month} className={inPeriod(m.month) ? "xd-col" : "xd-col muted"}>
                            <span className={diff >= 0 ? "xd-col-value mono xd-down" : "xd-col-value mono xd-up"}>
                                {empty ? "" : `${diff >= 0 ? "+" : "−"}${formatCompact(Math.abs(diff))}`}
                            </span>
                            <div className="xd-pair">
                                <div style={{ height: px(m.income, max), background: "#2f49b0" }} />
                                <div style={{ height: px(m.total, max), background: "#e8590c" }} />
                            </div>
                            <span className="xd-col-label">{monthShort(m.month, i === 0)}</span>
                        </div>
                    );
                })}
            </MonthChart>
            <div className="xd-total-line">
                <span className="row-meta">Net, last 12 months</span>
                <span className={net >= 0 ? "mono xd-down" : "mono xd-up"}>
                    {net >= 0 ? "+" : "−"}{formatWhole(Math.abs(net))}
                </span>
            </div>
        </section>
    );
}

export function LargestReceipts({ receipts }) {
    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>Largest expenses</h5>
                <span className="row-meta">Top {receipts.length} receipts in the period</span>
            </div>
            <div className="xd-receipt xd-receipt-head">
                <span>Date</span>
                <span className="num">Amount</span>
                <span>Group</span>
                <span>Lines</span>
                <span className="num">Items</span>
            </div>
            {receipts.map((r) => (
                <div key={r.expense_id} className="xd-receipt">
                    <span className="row-meta">{formatISODate(r.occurred_on)}</span>
                    <span className="num mono xd-strong">{formatWhole(r.total)}</span>
                    <span>{r.group}</span>
                    <span className="xd-ellipsis" title={r.summary}>{r.summary}</span>
                    <span className="num row-meta">{r.lines} {r.lines === 1 ? "line" : "lines"}</span>
                </div>
            ))}
            {receipts.length === 0 && <p className="xd-note">No expenses in this period.</p>}
        </section>
    );
}
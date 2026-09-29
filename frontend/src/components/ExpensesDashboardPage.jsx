import { useState, useEffect } from "react";
import { getExpenseDashboard, getExpenseGroups } from "../api/client";
import { currentMonth, shiftMonth, monthRange, formatISODate } from "./expenseUtils";
import { periodRange, previousLabel, stepMonths, GROUP_COLORS, PLAIN_BAR_COLOR } from "./dashboardUtils";
import {
    KpiRow, GroupsBreakdown, CategoriesPanel, TrendChart, FixedVariableChart, CashflowChart, LargestReceipts,
} from "./ExpenseDashboardBlocks";

const MODES = [
    ["month", "Month"],
    ["quarter", "Quarter"],
    ["year", "Year"],
    ["custom", "Custom"],
];

function ExpensesDashboardPage() {
    const [mode, setMode] = useState("month");
    const [month, setMonth] = useState(currentMonth());
    const [custom, setCustom] = useState(() => {
        const [from, to] = monthRange(currentMonth());
        return { from, to };
    });
    const [compare, setCompare] = useState(true);
    const [data, setData] = useState(null);
    const [groupNames, setGroupNames] = useState({});
    const [selectedId, setSelectedId] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const range = mode === "custom"
        ? { ...custom, label: `${formatISODate(custom.from)} – ${formatISODate(custom.to)}` }
        : periodRange(mode, month);
    const rangeValid = Boolean(range.from && range.to && range.to >= range.from);

    useEffect(() => {
        getExpenseGroups()
            .then((groups) => {
                const names = {};
                for (const g of groups) names[g.id] = g.name;
                setGroupNames(names);
            })
            .catch((err) => setError(err.message));
    }, []);

    useEffect(() => {
        if (!rangeValid) return;
        let ignore = false;
        setLoading(true);
        setError(null);
        getExpenseDashboard(range.from, range.to)
            .then((d) => {
                if (!ignore) {
                    setData(d);
                    setLoading(false);
                }
            })
            .catch((err) => {
                if (!ignore) {
                    setError(err.message);
                    setLoading(false);
                }
            });
        return () => {
            ignore = true;
        };
    }, [range.from, range.to, rangeValid]);

    let topIds = [];
    if (data) {
        const sums = {};
        for (const m of data.trend) {
            for (const g of m.by_group) sums[g.group_id] = (sums[g.group_id] || 0) + Number(g.total);
        }
        topIds = Object.entries(sums)
            .sort((a, b) => b[1] - a[1])
            .slice(0, GROUP_COLORS.length)
            .map(([id]) => id);
    }
    const colorOf = (id) => {
        const i = topIds.indexOf(id);
        return i >= 0 ? GROUP_COLORS[i] : PLAIN_BAR_COLOR;
    };

    const periodMonths = [range.from.slice(0, 7), range.to.slice(0, 7)];
    const inPeriod = (key) => key >= periodMonths[0] && key <= periodMonths[1];

    const activeGroups = data ? data.groups.filter((g) => Number(g.total) > 0) : [];
    const selected = activeGroups.find((g) => g.group_id === selectedId) || activeGroups[0];
    const prevLabel = data ? previousLabel(mode, data.previous) : "";
    const coverage = data ? data.coverage : null;

    return (
        <div className={loading && data ? "xd-page xd-loading" : "xd-page"}>
            <div className="page-header xd-header">
                <div>
                    <h1>Expenses dashboard</h1>
                    <p>Where the money goes{data ? `, in ${data.currency}` : ""}.</p>
                </div>
                <div className="xd-controls">
                    <div className="xd-segmented">
                        {MODES.map(([key, label]) => (
                            <button
                                type="button"
                                key={key}
                                className={mode === key ? "active" : ""}
                                onClick={() => setMode(key)}
                            >
                                {label}
                            </button>
                        ))}
                    </div>

                    {mode === "custom" ? (
                        <div className="xd-custom">
                            <input className="input" type="date" value={custom.from} onChange={(e) => setCustom({ ...custom, from: e.target.value })} />
                            <span className="row-meta">–</span>
                            <input className="input" type="date" value={custom.to} onChange={(e) => setCustom({ ...custom, to: e.target.value })} />
                        </div>
                    ) : (
                        <div className="month-nav">
                            <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, -stepMonths[mode]))}>←</button>
                            <h5>{range.label}</h5>
                            <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, stepMonths[mode]))}>→</button>
                        </div>
                    )}

                    <label className="checkbox">
                        <input type="checkbox" checked={compare} onChange={(e) => setCompare(e.target.checked)} />
                        Compare with previous period
                    </label>
                </div>
            </div>

            {error && <div className="error">{error}</div>}
            {!rangeValid && <div className="error">The end of the period is before its start.</div>}
            {!data && !error && <div className="loading">Loading…</div>}

            {data && (
                <>
                    <KpiRow summary={data.summary} compare={compare} prevLabel={prevLabel} />

                    <div className="xd-row-groups">
                        <GroupsBreakdown
                            groups={data.groups}
                            total={Number(data.summary.total) || 1}
                            selectedId={selected ? selected.group_id : null}
                            onSelect={setSelectedId}
                            colorOf={colorOf}
                            compare={compare}
                            prevLabel={prevLabel}
                        />
                        <CategoriesPanel group={selected} color={selected ? colorOf(selected.group_id) : PLAIN_BAR_COLOR} />
                    </div>

                    <TrendChart
                        trend={data.trend}
                        topIds={topIds}
                        colorOf={colorOf}
                        groupNames={groupNames}
                        inPeriod={inPeriod}
                    />

                    <div className="xd-row-halves">
                        <FixedVariableChart trend={data.trend} inPeriod={inPeriod} />
                        <CashflowChart trend={data.trend} inPeriod={inPeriod} />
                    </div>

                    <LargestReceipts receipts={data.largest} />

                    {coverage && (coverage.unconverted_days > 0 || coverage.unconverted_incomes > 0) && (
                        <p className="xd-note">
                            Not converted for lack of an exchange rate: {coverage.unconverted_days} days of expenses,{" "}
                            {coverage.unconverted_incomes} incomes. Totals above leave them out.
                        </p>
                    )}
                </>
            )}
        </div>
    );
}

export default ExpensesDashboardPage;
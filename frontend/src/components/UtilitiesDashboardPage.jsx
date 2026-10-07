import { useState, useEffect } from "react";
import { getUtilityDashboard } from "../api/client";
import { currentMonth, shiftMonth } from "./expenseUtils";
import { periodRange, previousLabel, stepMonths } from "./dashboardUtils";
import { ServiceCards, ConsumptionChart, EstimatedVsPaid, MonthTable, CoverageNote } from "./UtilityDashboardBlocks";

const MODES = [
    ["month", "Month"],
    ["quarter", "Quarter"],
    ["year", "Year"],
];

function UtilitiesDashboardPage() {
    const [mode, setMode] = useState("month");
    const [month, setMonth] = useState(shiftMonth(currentMonth(), -1));
    const [compare, setCompare] = useState(true);
    const [costService, setCostService] = useState(null);
    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const range = periodRange(mode, month);
    const from = range.from.slice(0, 7);
    const to = range.to.slice(0, 7);

    useEffect(() => {
        let ignore = false;
        setLoading(true);
        setError(null);
        getUtilityDashboard(from, to)
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
    }, [from, to]);

    const inPeriod = (key) => key >= from && key <= to;
    const prevLabel = data ? previousLabel(mode, { from: `${data.previous.from}-01`, to: `${data.previous.to}-01` }) : "";

    return (
        <div className={loading && data ? "xd-page xd-loading" : "xd-page"}>
            <div className="page-header xd-header">
                <div>
                    <h1>Utilities dashboard</h1>
                    <p>Consumption and estimated cost.</p>
                </div>
                <div className="xd-controls">
                    <div className="xd-segmented">
                        {MODES.map(([key, label]) => (
                            <button type="button" key={key} className={mode === key ? "active" : ""} onClick={() => setMode(key)}>
                                {label}
                            </button>
                        ))}
                    </div>
                    <div className="month-nav">
                        <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, -stepMonths[mode]))}>←</button>
                        <h5>{range.label}</h5>
                        <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, stepMonths[mode]))}>→</button>
                    </div>
                    <label className="checkbox">
                        <input type="checkbox" checked={compare} onChange={(e) => setCompare(e.target.checked)} />
                        Compare with previous period
                    </label>
                </div>
            </div>

            {error && <div className="error">{error}</div>}
            {!data && !error && <div className="loading">Loading…</div>}

            {data && data.services.length === 0 && (
                <p className="xd-note">No utility accounts yet. Add them on the Meters page.</p>
            )}

            {data && data.services.length > 0 && (
                <>
                    <ServiceCards services={data.services} currency={data.currency} compare={compare} prevLabel={prevLabel} />

                    <div className="xd-row-halves">
                        {data.services.map((s) => (
                            <ConsumptionChart key={s.service} block={s} months={data.months} inPeriod={inPeriod} />
                        ))}
                        <EstimatedVsPaid
                            services={data.services}
                            selected={costService}
                            onSelect={setCostService}
                            months={data.months}
                            inPeriod={inPeriod}
                        />
                    </div>

                    <MonthTable services={data.services} months={data.months} inPeriod={inPeriod} />

                    <CoverageNote coverage={data.coverage} />
                    <p className="xd-note">
                        A billing month's cost uses the tariff in force on its last day; the tier is picked by the
                        account's monthly total. Amounts in {data.currency}.
                    </p>
                </>
            )}
        </div>
    );
}

export default UtilitiesDashboardPage;
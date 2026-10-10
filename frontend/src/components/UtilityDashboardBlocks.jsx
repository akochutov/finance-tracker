import { formatWhole, formatCompact, percentChange, niceMax, monthShort } from "./dashboardUtils";

const CHART_HEIGHT = 200;

const PALETTES = {
    electricity: ["#2f49b0", "#91a7ff", "#4c6ef5", "#bac8ff"],
    gas: ["#e8590c", "#ffa94d", "#d9480f", "#ffc078"],
    water: ["#0b7285", "#66d9e8", "#1098ad", "#99e9f2"],
};
const FALLBACK = ["#5c7cfa", "#20c997", "#fab005", "#e64980"];

export function seriesColor(service, i) {
    const palette = PALETTES[service] || FALLBACK;
    return palette[i % palette.length];
}

const num = (v) => (v === null || v === undefined ? null : Number(v));

function formatAxis(v) {
    return v >= 10 ? formatCompact(v) : Number(v.toFixed(1)).toString();
}

function formatQty(v) {
    return Number(v).toLocaleString(undefined, { maximumFractionDigits: 1 });
}

function paidVerdict(paid, estimated) {
    const diff = paid - estimated;
    if (estimated > 0 && Math.abs(diff) <= estimated * 0.02) return { text: "matches the estimate", cls: "ud-ok" };
    return diff > 0
        ? { text: `overpaid ${formatWhole(diff)}`, cls: "xd-up" }
        : { text: `underpaid ${formatWhole(-diff)}`, cls: "xd-down" };
}

export function ServiceCards({ services, currency, compare, prevLabel }) {
    return (
        <div className="ud-cards">
            {services.map((s) => (
                <ServiceCard key={s.service} block={s} currency={currency} compare={compare} prevLabel={prevLabel} />
            ))}
        </div>
    );
}

function ServiceCard({ block, currency, compare, prevLabel }) {
    const sum = block.summary;
    const prev = block.previous;
    const estimated = Number(sum.estimated);
    const paid = Number(sum.paid);
    const change = percentChange(estimated, prev.estimated);
    const nothingToCompare = estimated === 0 && Number(prev.estimated) === 0;
    const color = seriesColor(block.service, 0);

    const consumption = sum.zones.length > 0
        ? sum.zones.map((z) => `${z.zone} ${formatQty(z.quantity)}`).join(" · ") + ` ${block.unit}`
        : `${formatQty(sum.consumption)} ${block.unit}`;

    return (
        <div className="xd-panel">
            <div className="xd-kpi-label ud-card-label">
                <i style={{ background: color }} />
                {block.name} · estimated cost
            </div>
            <div className="xd-kpi-value mono">
                {formatWhole(estimated)} <span className="ud-currency">{currency}</span>
            </div>
            <div className="xd-kpi-sub">{consumption}</div>

            {sum.tracks_payments ? (
                <div className="xd-kpi-sub">
                    Paid <span className="mono">{formatWhole(paid)}</span>
                    {estimated > 0 && (
                        <>
                            {" · "}
                            <span className={paidVerdict(paid, estimated).cls}>{paidVerdict(paid, estimated).text}</span>
                        </>
                    )}
                </div>
            ) : (
                <div className="xd-kpi-sub row-meta">Payments not tracked: set a category on the account.</div>
            )}

            {sum.months_without_tariff > 0 && (
                <div className="xd-kpi-sub xd-up">
                    {sum.months_without_tariff} {sum.months_without_tariff === 1 ? "month" : "months"} without a tariff
                </div>
            )}

            {compare && (
                <div className="xd-kpi-sub">
                    {nothingToCompare ? (
                        <span className="row-meta">no cost in {prevLabel} either</span>
                    ) : change === null ? (
                        <span className="row-meta">new vs {prevLabel}</span>
                    ) : (
                        <>
                            <span className={change >= 0 ? "xd-up" : "xd-down"}>
                                {change >= 0 ? "▲" : "▼"} {Math.abs(change).toFixed(1)}%
                            </span>{" "}
                            vs {prevLabel} · {formatWhole(prev.estimated)}
                        </>
                    )}
                </div>
            )}
        </div>
    );
}

export function GroupedBars({ title, months, series, inPeriod, note, extra }) {
    const max = niceMax(Math.max(0, ...series.flatMap((s) => s.values.map((v) => v ?? 0))));
    const width = series.length <= 1 ? 18 : series.length === 2 ? 11 : 8;

    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <div className="ud-chart-title">
                    <h5>{title}</h5>
                    {extra}
                </div>
                <div className="xd-legend">
                    {series.map((s) => (
                        <span key={s.key}><i style={{ background: s.color }} />{s.label}</span>
                    ))}
                </div>
            </div>

            <div className="ud-chart">
                <div className="ud-axis">
                    <span>{formatAxis(max)}</span>
                    <span>{formatAxis(max / 2)}</span>
                    <span>0</span>
                </div>
                <div className="ud-plot">
                    {months.map((m, i) => (
                        <div key={m} className={inPeriod(m) ? "ud-col" : "ud-col faded"}>
                            <div className="ud-bars">
                                {series.map((s) => (
                                    <div
                                        key={s.key}
                                        className="ud-bar"
                                        title={`${s.label} · ${monthShort(m, true)}: ${s.values[i] === null ? "—" : formatQty(s.values[i])}`}
                                        style={{
                                            width: `${width}px`,
                                            height: `${((s.values[i] ?? 0) / max) * CHART_HEIGHT}px`,
                                            background: s.color,
                                        }}
                                    />
                                ))}
                            </div>
                            <span className="ud-month">{monthShort(m, i === 0)}</span>
                        </div>
                    ))}
                </div>
            </div>

            {note && <p className="xd-note">{note}</p>}
        </section>
    );
}

export function ConsumptionChart({ block, months, inPeriod }) {
    const series = block.series.map((s, i) => ({
        key: s.key,
        label: s.label,
        color: seriesColor(block.service, i),
        values: s.consumption.map(num),
    }));
    const zones = block.series.some((s) => s.kind === "zone");
    const note = zones
        ? "Day and night from the meter's registers."
        : block.series.length > 1
            ? "One column per meter."
            : null;

    return (
        <GroupedBars
            title={`${block.name}, ${block.unit}`}
            months={months}
            series={series}
            inPeriod={inPeriod}
            note={note}
        />
    );
}

export function EstimatedVsPaid({ services, selected, onSelect, months, inPeriod }) {
    const block = services.find((s) => s.service === selected) || services[0];
    if (!block) return null;

    const pills = (
        <div className="ud-pills">
            {services.map((s) => (
                <button
                    key={s.service}
                    type="button"
                    className={s.service === block.service ? "rd-tab active" : "rd-tab"}
                    onClick={() => onSelect(s.service)}
                >
                    {s.name}
                </button>
            ))}
        </div>
    );

    const series = [
        { key: "estimated", label: "Estimated", color: "#3b5bdb", values: block.estimated.map(num) },
        { key: "paid", label: "Paid", color: "#f59f00", values: block.paid.map(num) },
    ];

    return (
        <GroupedBars
            title="Estimated vs paid"
            extra={pills}
            months={months}
            series={series}
            inPeriod={inPeriod}
            note={
                block.previous.tracks_payments || block.summary.tracks_payments
                    ? "Paid: expense lines of the account's category, by the month of their period."
                    : "Payments are not tracked for this service: set a category on its account."
            }
        />
    );
}

function monthlyConsumption(block, i) {
    let sum = null;
    for (const s of block.series) {
        const v = s.consumption[i];
        if (v === null || v === undefined) continue;
        sum = (sum ?? 0) + Number(v);
    }
    return sum;
}

export function MonthTable({ services, months, inPeriod }) {
    const rows = months.map((m, i) => ({ m, i })).reverse();

    return (
        <section className="xd-panel">
            <div className="xd-panel-head">
                <h5>By month</h5>
                <span className="row-meta">consumption · estimated cost</span>
            </div>
            <div className="ud-table-wrap">
                <table className="ud-table">
                    <thead>
                        <tr>
                            <th>Month</th>
                            {services.map((s) => (
                                <th key={s.service} className="num">{s.name}</th>
                            ))}
                        </tr>
                    </thead>
                    <tbody>
                        {rows.map(({ m, i }) => (
                            <tr key={m} className={inPeriod(m) ? "current" : ""}>
                                <td>{monthShort(m, true)}</td>
                                {services.map((s) => {
                                    const q = monthlyConsumption(s, i);
                                    const cost = s.estimated[i];
                                    return (
                                        <td key={s.service} className="num mono">
                                            {q === null ? "—" : `${formatQty(q)} ${s.unit}`}
                                            {q !== null && (
                                                <span className="ud-cost">{cost === null ? "no cost" : formatWhole(cost)}</span>
                                            )}
                                        </td>
                                    );
                                })}
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </section>
    );
}

export function CoverageNote({ coverage }) {
    const parts = [];
    if (coverage.missing_tariffs.length) parts.push(`no tariff: ${coverage.missing_tariffs.join(", ")}`);
    if (coverage.missing_rates.length) parts.push(`no exchange rate: ${coverage.missing_rates.join(", ")}`);
    if (parts.length === 0) return null;
    return <p className="xd-note">Incomplete — {parts.join("; ")}.</p>;
}
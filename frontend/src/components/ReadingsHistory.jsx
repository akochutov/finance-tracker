import { useState, useEffect } from "react";
import { getReadings, updateReading, deleteReading } from "../api/client";
import { formatISODate } from "./expenseUtils";
import { isDayNight, zonesOf, zoneLabel, meterTitle, groupByDate, billingMonthKey, monthLabel, formatReading } from "./meterUtils";

function ReadingsHistory({ meters, accounts, serviceTypes, selectedId, onSelect, reloadKey, onChanged }) {
    const [readings, setReadings] = useState([]);
    const [localKey, setLocalKey] = useState(0);
    const [error, setError] = useState(null);
    const [editRow, setEditRow] = useState(null);
    const [showRemoved, setShowRemoved] = useState(false);

    const active = meters.filter((m) => !m.removed_on);
    const removed = meters.filter((m) => m.removed_on);
    const meter = meters.find((m) => m.id === selectedId) || active[0] || meters[0] || null;

    useEffect(() => {
        if (!meter) return;
        let ignore = false;
        setError(null);
        setEditRow(null);
        getReadings(meter.id)
            .then((rds) => {
                if (!ignore) setReadings(rds);
            })
            .catch((err) => {
                if (!ignore) setError(err.message);
            });
        return () => {
            ignore = true;
        };
    }, [meter?.id, reloadKey, localKey]);

    if (!meter) return null;

    const dayNight = isDayNight(meter);
    const zones = zonesOf(meter);
    const rounds = groupByDate(readings);

    async function run(action) {
        setError(null);
        try {
            await action();
            setLocalKey((k) => k + 1);
            onChanged();
            return true;
        } catch (err) {
            setError(err.message);
            setLocalKey((k) => k + 1);
            return false;
        }
    }

    function startEdit(round) {
        const vals = {};
        for (const z of zones) {
            const r = round.byZone[z];
            vals[z] = r ? String(r.value) : "";
        }
        setEditRow({ key: round.taken_on, round, date: round.taken_on.slice(0, 10), values: vals });
    }

    async function saveEdit() {
        const ok = await run(async () => {
            for (const z of zones) {
                const r = editRow.round.byZone[z];
                if (!r) continue;
                await updateReading(meter.id, r.id, { taken_on: editRow.date, value: editRow.values[z] });
            }
        });
        if (ok) setEditRow(null);
    }

    async function remove(round) {
        if (!window.confirm(`Delete the reading of ${formatISODate(round.taken_on)}?`)) return;
        await run(async () => {
            for (const r of Object.values(round.byZone)) {
                await deleteReading(meter.id, r.id);
            }
        });
    }

    const rowClass = dayNight ? "rd-hist-row day-night" : "rd-hist-row";
    const tabs = showRemoved || meter.removed_on ? [...active, ...removed] : active;

    function deltaCell(value, key) {
        return (
            <span key={key} className={value === null ? "num mt-delta none" : "num mono mt-delta"}>
                {value === null ? "—" : `+${formatReading(value)}`}
            </span>
        );
    }

    return (
        <section className="xd-panel">
            <div className="rd-hist-head">
                <div className="mt-section-title">History</div>
                <div className="rd-tabs">
                    {tabs.map((m) => {
                        const acc = accounts.find((a) => a.id === m.account_id);
                        const t = acc ? serviceTypes.find((x) => x.code === acc.service) : null;
                        const classes = ["rd-tab"];
                        if (m.id === meter.id) classes.push("active");
                        if (m.removed_on) classes.push("removed");
                        return (
                            <button key={m.id} type="button" className={classes.join(" ")} onClick={() => onSelect(m.id)}>
                                {meterTitle(m)} · {t ? t.name.toLowerCase() : ""}
                            </button>
                        );
                    })}
                    {removed.length > 0 && !meter.removed_on && (
                        <button type="button" className="mt-link" onClick={() => setShowRemoved(!showRemoved)}>
                            {showRemoved ? "hide removed" : `show removed (${removed.length})`}
                        </button>
                    )}
                </div>
            </div>

            {error && <div className="error">{error}</div>}

            <div className={`${rowClass} rd-hist-labels`}>
                <span>Date</span>
                <span>For month</span>
                {dayNight ? (
                    <>
                        <span className="num">Day</span>
                        <span className="num">Night</span>
                        <span className="num">Day +</span>
                        <span className="num">Night +</span>
                        <span className="num">Total</span>
                    </>
                ) : (
                    <>
                        <span className="num">Meter</span>
                        <span className="num">Consumption</span>
                    </>
                )}
                <span></span>
            </div>

            {rounds.map((round, i) => {
                const older = rounds[i + 1] || null;
                const deltas = zones.map((z) => {
                    const r = round.byZone[z];
                    const o = older ? older.byZone[z] : null;
                    return r && o ? Number(r.value) - Number(o.value) : null;
                });
                const total = deltas.every((d) => d !== null) ? deltas.reduce((a, b) => a + b, 0) : null;
                const key = billingMonthKey(round.taken_on);

                if (editRow && editRow.key === round.taken_on) {
                    return (
                        <div key={round.taken_on} className={rowClass}>
                            <input className="input" type="date" value={editRow.date} onChange={(e) => setEditRow({ ...editRow, date: e.target.value })} />
                            <span className="row-meta">{round.is_initial ? "Initial" : monthLabel(billingMonthKey(editRow.date))}</span>
                            {zones.map((z) => (
                                <input
                                    key={z}
                                    className="input mono"
                                    type="number"
                                    step="any"
                                    min="0"
                                    placeholder={zoneLabel(z).toLowerCase()}
                                    value={editRow.values[z]}
                                    onChange={(e) => setEditRow({ ...editRow, values: { ...editRow.values, [z]: e.target.value } })}
                                />
                            ))}
                            <span></span>
                            {dayNight && (
                                <>
                                    <span></span>
                                    <span></span>
                                </>
                            )}
                            <span className="mt-row-actions">
                                <button type="button" className="mt-link" onClick={saveEdit}>save</button>
                                <button type="button" className="mt-link" onClick={() => setEditRow(null)}>cancel</button>
                            </span>
                        </div>
                    );
                }

                return (
                    <div key={round.taken_on} className={round.is_initial ? `${rowClass} initial` : rowClass}>
                        <span>{formatISODate(round.taken_on)}</span>
                        <span className="row-meta">{round.is_initial ? "Initial" : older ? monthLabel(key) : "—"}</span>
                        {zones.map((z) => (
                            <span key={z} className="num mono">
                                {round.byZone[z] ? formatReading(round.byZone[z].value) : "—"}
                            </span>
                        ))}
                        {round.is_initial ? (
                            <>
                                <span className="num mt-delta none">starting point</span>
                                {dayNight && (
                                    <>
                                        <span></span>
                                        <span></span>
                                    </>
                                )}
                            </>
                        ) : (
                            <>
                                {dayNight && deltas.map((d, k) => deltaCell(d, zones[k]))}
                                {deltaCell(total, "total")}
                            </>
                        )}
                        <span className="mt-row-actions">
                            <button type="button" className="mt-link" onClick={() => startEdit(round)}>edit</button>
                            {!round.is_initial && (
                                <button type="button" className="mt-link mt-link-danger" onClick={() => remove(round)}>delete</button>
                            )}
                        </span>
                    </div>
                );
            })}

            {rounds.length === 0 && <p className="xd-note">No readings yet.</p>}

            <p className="xd-note">
                {dayNight
                    ? "Each register counts on its own: day and night are the differences with the previous reading of the same register."
                    : "Consumption is the difference with the previous reading. A reading of early October closes September."}
            </p>
        </section>
    );
}

export default ReadingsHistory;
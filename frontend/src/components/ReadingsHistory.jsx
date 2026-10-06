import { useState, useEffect } from "react";
import { getReadings, getZoneUsage, updateReading, deleteReading, setZoneUsage, deleteZoneUsage } from "../api/client";
import { formatISODate } from "./expenseUtils";
import { isDayNight, billingMonthKey, monthLabel, formatReading, sameAmount } from "./meterUtils";

function ReadingsHistory({ meters, accounts, serviceTypes, selectedId, onSelect, reloadKey, onChanged }) {
    const [readings, setReadings] = useState([]);
    const [usage, setUsage] = useState([]);
    const [localKey, setLocalKey] = useState(0);
    const [error, setError] = useState(null);
    const [editRow, setEditRow] = useState(null);

    const meter = meters.find((m) => m.id === selectedId) || meters[0] || null;
    const account = meter ? accounts.find((a) => a.id === meter.account_id) : null;
    const type = account ? serviceTypes.find((t) => t.code === account.service) : null;
    const dayNight = isDayNight(account);
    const unit = type ? type.unit : "";

    useEffect(() => {
        if (!meter) return;
        let ignore = false;
        setError(null);
        setEditRow(null);
        Promise.all([getReadings(meter.id), dayNight ? getZoneUsage(account.id) : Promise.resolve([])])
            .then(([rds, zu]) => {
                if (ignore) return;
                setReadings(rds);
                setUsage(zu);
            })
            .catch((err) => {
                if (!ignore) setError(err.message);
            });
        return () => {
            ignore = true;
        };
    }, [meter?.id, dayNight, reloadKey, localKey]);

    if (!meter) return null;

    const splitBy = {};
    for (const u of usage) {
        const key = u.month.slice(0, 7);
        splitBy[key] = { ...(splitBy[key] || {}), [u.zone]: Number(u.quantity) };
    }

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

    function startEdit(r) {
        const key = billingMonthKey(r.taken_on);
        const s = splitBy[key] || {};
        setEditRow({
            id: r.id,
            date: r.taken_on.slice(0, 10),
            value: String(r.value),
            day: s.day !== undefined ? String(s.day) : "",
            night: s.night !== undefined ? String(s.night) : "",
            hadSplit: Boolean(splitBy[key]),
            isInitial: r.is_initial,
        });
    }

    async function saveEdit() {
        const ok = await run(async () => {
            await updateReading(meter.id, editRow.id, { taken_on: editRow.date, value: editRow.value });
            if (dayNight && !editRow.isInitial) {
                const key = billingMonthKey(editRow.date);
                if (editRow.day !== "" && editRow.night !== "") {
                    await setZoneUsage(account.id, key, { day: editRow.day, night: editRow.night });
                } else if (editRow.day === "" && editRow.night === "" && editRow.hadSplit) {
                    await deleteZoneUsage(account.id, key);
                }
            }
        });
        if (ok) setEditRow(null);
    }

    async function remove(r) {
        if (!window.confirm(`Delete the reading of ${formatISODate(r.taken_on)}?`)) return;
        await run(() => deleteReading(meter.id, r.id));
    }

    const rowClass = dayNight ? "rd-hist-row day-night" : "rd-hist-row";
    const tabs = [...meters].sort((a, b) => Number(Boolean(a.removed_on)) - Number(Boolean(b.removed_on)));

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
                                {m.serial} · {t ? t.name.toLowerCase() : ""}
                            </button>
                        );
                    })}
                </div>
            </div>

            {error && <div className="error">{error}</div>}

            <div className={`${rowClass} rd-hist-labels`}>
                <span>Date</span>
                <span>For month</span>
                <span className="num">Meter</span>
                <span className="num">Consumption</span>
                {dayNight && (
                    <>
                        <span className="num">Provider day</span>
                        <span className="num">Provider night</span>
                        <span className="num">Check</span>
                    </>
                )}
                <span></span>
            </div>

            {readings.map((r, i) => {
                const older = readings[i + 1] || null;
                const consumption = older ? Number(r.value) - Number(older.value) : null;
                const key = billingMonthKey(r.taken_on);
                const split = !r.is_initial && older ? splitBy[key] : null;

                if (editRow && editRow.id === r.id) {
                    return (
                        <div key={r.id} className={rowClass}>
                            <input className="input" type="date" value={editRow.date} onChange={(e) => setEditRow({ ...editRow, date: e.target.value })} />
                            <span className="row-meta">{r.is_initial ? "Initial" : monthLabel(billingMonthKey(editRow.date))}</span>
                            <input className="input mono" type="number" step="any" min="0" value={editRow.value} onChange={(e) => setEditRow({ ...editRow, value: e.target.value })} />
                            <span></span>
                            {dayNight && (
                                r.is_initial ? (
                                    <><span></span><span></span><span></span></>
                                ) : (
                                    <>
                                        <input className="input mono" type="number" step="any" min="0" placeholder="day" value={editRow.day} onChange={(e) => setEditRow({ ...editRow, day: e.target.value })} />
                                        <input className="input mono" type="number" step="any" min="0" placeholder="night" value={editRow.night} onChange={(e) => setEditRow({ ...editRow, night: e.target.value })} />
                                        <span></span>
                                    </>
                                )
                            )}
                            <span className="mt-row-actions">
                                <button type="button" className="mt-link" onClick={saveEdit}>save</button>
                                <button type="button" className="mt-link" onClick={() => setEditRow(null)}>cancel</button>
                            </span>
                        </div>
                    );
                }

                let check = <span></span>;
                if (split && split.day !== undefined && split.night !== undefined && consumption !== null) {
                    const sum = split.day + split.night;
                    check = sameAmount(sum, consumption)
                        ? <span className="num rd-check ok">✓ matches</span>
                        : <span className="num rd-check bad">{sum > consumption ? "+" : "−"}{formatReading(Math.abs(sum - consumption))} {unit}</span>;
                }

                return (
                    <div key={r.id} className={r.is_initial ? `${rowClass} initial` : rowClass}>
                        <span>{formatISODate(r.taken_on)}</span>
                        <span className="row-meta">{r.is_initial ? "Initial" : older ? monthLabel(key) : "—"}</span>
                        <span className="num mono">{formatReading(r.value)}</span>
                        <span className={consumption === null ? "num mt-delta none" : "num mono mt-delta"}>
                            {r.is_initial ? "starting point" : consumption === null ? "—" : `+${formatReading(consumption)}`}
                        </span>
                        {dayNight && (
                            <>
                                <span className="num mono">{split && split.day !== undefined ? formatReading(split.day) : ""}</span>
                                <span className="num mono">{split && split.night !== undefined ? formatReading(split.night) : ""}</span>
                                {check}
                            </>
                        )}
                        <span className="mt-row-actions">
                            <button type="button" className="mt-link" onClick={() => startEdit(r)}>edit</button>
                            {!r.is_initial && (
                                <button type="button" className="mt-link mt-link-danger" onClick={() => remove(r)}>delete</button>
                            )}
                        </span>
                    </div>
                );
            })}

            {readings.length === 0 && <p className="xd-note">No readings yet.</p>}

            <p className="xd-note">
                {dayNight
                    ? "The meter shows one value; the provider splits each month into day and night. The check compares the split with the meter."
                    : "Consumption is the difference with the previous reading. A reading of early October closes September."}
            </p>
        </section>
    );
}

export default ReadingsHistory;
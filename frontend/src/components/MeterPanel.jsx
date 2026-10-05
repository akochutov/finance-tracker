import { useState, useEffect } from "react";
import { getReadings, createReadings, updateReading, deleteReading } from "../api/client";
import { formatISODate, todayISO } from "./expenseUtils";
import { meterZones, zoneLabel, isDual, groupReadings, consumption, billingMonth, formatReading } from "./meterUtils";

function MeterPanel({ meter, account, serviceType, actions }) {
    const zones = meterZones(meter);
    const dual = isDual(meter);
    const unit = serviceType ? serviceType.unit : "";

    const [readings, setReadings] = useState([]);
    const [reloadKey, setReloadKey] = useState(0);
    const [error, setError] = useState(null);

    const [editing, setEditing] = useState(false);
    const [serial, setSerial] = useState(meter.serial);
    const [installedOn, setInstalledOn] = useState(meter.installed_on.slice(0, 10));
    const [removedOn, setRemovedOn] = useState(meter.removed_on ? meter.removed_on.slice(0, 10) : "");

    const [takenOn, setTakenOn] = useState(todayISO());
    const [values, setValues] = useState({});

    const [editRow, setEditRow] = useState(null);

    useEffect(() => {
        let ignore = false;
        getReadings(meter.id)
            .then((data) => {
                if (!ignore) setReadings(data);
            })
            .catch((err) => {
                if (!ignore) setError(err.message);
            });
        return () => {
            ignore = true;
        };
    }, [meter.id, reloadKey]);

    async function run(action) {
        setError(null);
        try {
            await action();
            setReloadKey((k) => k + 1);
            return true;
        } catch (err) {
            setError(err.message);
            setReloadKey((k) => k + 1);
            return false;
        }
    }

    async function saveMeter(e) {
        e.preventDefault();
        const ok = await actions.updateMeter(meter.id, {
            serial,
            installed_on: installedOn,
            removed_on: removedOn || null,
        });
        if (ok) setEditing(false);
    }

    async function removeMeter() {
        if (window.confirm(`Delete meter ${meter.serial} with all its readings? This cannot be undone.`)) {
            await actions.deleteMeter(meter.id);
        }
    }

    async function addReading(e) {
        e.preventDefault();
        const body = { taken_on: takenOn, values: {} };
        for (const z of zones) body.values[z] = values[z] ?? "";
        if (await run(() => createReadings(meter.id, body))) setValues({});
    }

    async function saveRow(row) {
        const ok = await run(async () => {
            for (const z of zones) {
                const reading = row.byZone[z];
                if (!reading) continue;
                await updateReading(meter.id, reading.id, { taken_on: editRow.date, value: editRow.values[z] });
            }
        });
        if (ok) setEditRow(null);
    }

    async function removeRow(row) {
        if (!window.confirm(`Delete the reading of ${formatISODate(row.date)}?`)) return;
        await run(async () => {
            for (const z of zones) {
                const reading = row.byZone[z];
                if (reading) await deleteReading(meter.id, reading.id);
            }
        });
    }

    function startEditRow(row) {
        const vals = {};
        for (const z of zones) vals[z] = row.byZone[z] ? String(row.byZone[z].value) : "";
        setEditRow({ key: row.date, date: row.date, values: vals });
    }

    const rows = groupReadings(meter, readings);
    const rowClass = dual ? "mt-reading-row dual" : "mt-reading-row";

    return (
        <section className="xd-panel">
            <div className="mt-panel-head">
                <div>
                    <div className="mt-caption">
                        {serviceType ? serviceType.name : account.service} · account {account.number}
                    </div>
                    <div className="mt-title">Meter <span className="mono">{meter.serial}</span></div>
                    <div className="mt-pills">
                        <span className="mt-pill accent">{dual ? "Day / night" : "Single"}</span>
                        <span className="mt-pill">Installed {formatISODate(meter.installed_on)}</span>
                        {meter.removed_on && <span className="mt-pill">Removed {formatISODate(meter.removed_on)}</span>}
                        <span className="mt-pill">Unit: {unit}</span>
                    </div>
                </div>
                {!editing && (
                    <div className="mt-actions">
                        <button className="btn btn-secondary btn-sm" onClick={() => setEditing(true)}>Edit</button>
                        <button className="btn btn-ghost btn-sm" onClick={removeMeter}>Delete</button>
                    </div>
                )}
            </div>

            {editing && (
                <form className="mt-section" onSubmit={saveMeter}>
                    <div className="mt-section-title">Edit meter</div>
                    <div className="mt-reading-form">
                        <div className="field">
                            <label>Serial</label>
                            <input className="input" value={serial} onChange={(e) => setSerial(e.target.value)} />
                        </div>
                        <div className="field">
                            <label>Installed</label>
                            <input className="input" type="date" value={installedOn} onChange={(e) => setInstalledOn(e.target.value)} />
                        </div>
                        <div className="field">
                            <label>Removed</label>
                            <input className="input" type="date" value={removedOn} onChange={(e) => setRemovedOn(e.target.value)} />
                        </div>
                        <button type="submit" className="btn btn-primary btn-sm">Save</button>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={() => setEditing(false)}>Cancel</button>
                    </div>
                    <div className="mt-hint">Leave “Removed” empty while the meter is in place. Readings must stay within these dates.</div>
                </form>
            )}

            {error && <div className="error" style={{ marginTop: 16 }}>{error}</div>}

            {!meter.removed_on && (
                <form className="mt-section" onSubmit={addReading}>
                    <div className="mt-section-title">Add reading</div>
                    <div className="mt-reading-form">
                        <div className="field">
                            <label>Date</label>
                            <input className="input" type="date" value={takenOn} onChange={(e) => setTakenOn(e.target.value)} />
                        </div>
                        {zones.map((z) => (
                            <div className="field" key={z}>
                                <label>{zoneLabel(z)}, {unit}</label>
                                <input
                                    className="input mono"
                                    type="number"
                                    step="any"
                                    min="0"
                                    value={values[z] ?? ""}
                                    onChange={(e) => setValues({ ...values, [z]: e.target.value })}
                                />
                            </div>
                        ))}
                        <button type="submit" className="btn btn-primary btn-sm">Save</button>
                    </div>
                    <div className="mt-hint">
                        {dual ? "Both dials are saved together. " : ""}
                        A value below the previous reading is rejected.
                    </div>
                </form>
            )}

            <div className="mt-readings">
                <div className={`${rowClass} mt-reading-head`}>
                    <span>Date</span>
                    <span>For month</span>
                    {zones.map((z) => (
                        <ZoneHead key={z} zone={z} dual={dual} />
                    ))}
                    <span></span>
                </div>

                {rows.map((row, i) =>
                    editRow && editRow.key === row.date ? (
                        <div key={row.date} className={rowClass}>
                            <input
                                className="input"
                                type="date"
                                value={editRow.date}
                                onChange={(e) => setEditRow({ ...editRow, date: e.target.value })}
                            />
                            <span></span>
                            {zones.map((z) => (
                                <EditCells
                                    key={z}
                                    value={editRow.values[z]}
                                    onChange={(v) => setEditRow({ ...editRow, values: { ...editRow.values, [z]: v } })}
                                />
                            ))}
                            <span className="mt-row-actions">
                                <button type="button" className="mt-link" onClick={() => saveRow(row)}>save</button>
                                <button type="button" className="mt-link" onClick={() => setEditRow(null)}>cancel</button>
                            </span>
                        </div>
                    ) : (
                        <div key={row.date} className={rowClass}>
                            <span>{formatISODate(row.date)}</span>
                            <span className="row-meta">{i < rows.length - 1 ? billingMonth(row.date) : "—"}</span>
                            {zones.map((z) => (
                                <ValueCells key={z} reading={row.byZone[z]} delta={consumption(rows, i, z)} />
                            ))}
                            <span className="mt-row-actions">
                                <button type="button" className="mt-link" onClick={() => startEditRow(row)}>edit</button>
                                <button type="button" className="mt-link mt-link-danger" onClick={() => removeRow(row)}>delete</button>
                            </span>
                        </div>
                    )
                )}

                {rows.length === 0 && <p className="xd-note">No readings yet. The first one is the starting point; consumption appears from the second.</p>}
            </div>

            <p className="xd-note">
                Consumption is the difference with the previous reading. A reading taken in early October closes September — the month of the bill.
            </p>
        </section>
    );
}

function ZoneHead({ zone, dual }) {
    return (
        <>
            <span className="num">{dual ? zoneLabel(zone) : "Value"}</span>
            <span className="num">{dual ? `Δ ${zoneLabel(zone).toLowerCase()}` : "Consumption"}</span>
        </>
    );
}

function ValueCells({ reading, delta }) {
    return (
        <>
            <span className="num mono">{reading ? formatReading(reading.value) : "—"}</span>
            <span className={delta === null ? "num mono mt-delta none" : "num mono mt-delta"}>
                {delta === null ? "—" : `+${formatReading(delta)}`}
            </span>
        </>
    );
}

function EditCells({ value, onChange }) {
    return (
        <>
            <input className="input mono" type="number" step="any" min="0" value={value} onChange={(e) => onChange(e.target.value)} />
            <span></span>
        </>
    );
}

export default MeterPanel;
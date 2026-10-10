import { useState, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import {
    getServiceTypes, getAddresses, getUtilityAccounts, getMeters, getLatestReadings,
    createReadings,
} from "../api/client";
import { todayISO, formatISODate } from "./expenseUtils";
import { zonesOf, zoneLabel, meterTitle, registerKey, billingMonthKey, monthLabel, formatReading } from "./meterUtils";
import ReadingsHistory from "./ReadingsHistory";

function ReadingsPage() {
    const [serviceTypes, setServiceTypes] = useState([]);
    const [addresses, setAddresses] = useState([]);
    const [accounts, setAccounts] = useState([]);
    const [meters, setMeters] = useState([]);
    const [latest, setLatest] = useState([]);
    const [loaded, setLoaded] = useState(false);
    const [error, setError] = useState(null);
    const [status, setStatus] = useState(null);
    const [historyKey, setHistoryKey] = useState(0);

    const [takenOn, setTakenOn] = useState(todayISO());
    const [values, setValues] = useState({});

    const [searchParams, setSearchParams] = useSearchParams();
    const selectedId = searchParams.get("meter");

    async function loadAll() {
        try {
            const [types, addrs, accs, mts, last] = await Promise.all([
                getServiceTypes(), getAddresses(), getUtilityAccounts(), getMeters(), getLatestReadings(),
            ]);
            setServiceTypes(types);
            setAddresses(addrs);
            setAccounts(accs);
            setMeters(mts);
            setLatest(last);
            setLoaded(true);
        } catch (err) {
            setError(err.message);
        }
    }

    useEffect(() => {
        loadAll();
    }, []);

    const typeOf = (code) => serviceTypes.find((t) => t.code === code);
    const latestBy = Object.fromEntries(latest.map((r) => [registerKey(r.meter_id, r.zone), r]));
    const month = billingMonthKey(takenOn);

    const activeAddress = new Set(addresses.filter((a) => a.is_active).map((a) => a.id));
    const groups = accounts
        .filter((acc) => acc.is_active && activeAddress.has(acc.address_id))
        .map((acc) => ({
            account: acc,
            type: typeOf(acc.service),
            meters: meters.filter((m) => m.account_id === acc.id && !m.removed_on),
        }))
        .filter((g) => g.meters.length > 0);

    const roundMeters = groups.flatMap((g) => g.meters);

    function consumptionOf(key) {
        const raw = values[key];
        const prev = latestBy[key];
        if (raw === undefined || raw === "" || !prev) return null;
        return Number(raw) - Number(prev.value);
    }

    async function save(e) {
        e.preventDefault();
        setError(null);
        setStatus(null);

        const readings = [];
        const incomplete = [];
        for (const m of roundMeters) {
            const zones = zonesOf(m);
            const filled = zones.filter((z) => (values[registerKey(m.id, z)] ?? "") !== "");
            if (filled.length === 0) continue;
            if (filled.length < zones.length) {
                incomplete.push(meterTitle(m));
                continue;
            }
            for (const z of zones) {
                readings.push({ meter_id: m.id, zone: z, value: values[registerKey(m.id, z)] });
            }
        }

        if (incomplete.length > 0) {
            setError(`Enter both day and night for ${incomplete.join(", ")}.`);
            return;
        }
        if (readings.length === 0) {
            setError("Nothing to save: enter at least one value.");
            return;
        }

        const meterCount = new Set(readings.map((r) => r.meter_id)).size;
        try {
            await createReadings({ taken_on: takenOn, readings });
            setValues({});
            setStatus(`Saved readings of ${meterCount} ${meterCount === 1 ? "meter" : "meters"}.`);
            setHistoryKey((k) => k + 1);
            await loadAll();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <div className="xd-page">
            <div className="page-header">
                <h1>Readings</h1>
                <p>Monthly meter readings.</p>
            </div>

            {error && <div className="error">{error}</div>}
            {!loaded && !error && <div className="loading">Loading…</div>}

            {loaded && (
                <>
                    <form className="xd-panel" onSubmit={save}>
                        <div className="rd-form-head">
                            <div>
                                <div className="mt-section-title">Enter readings</div>
                                <div className="mt-hint">
                                    All active meters at once. Readings of {formatISODate(takenOn)} close {monthLabel(month)}.
                                </div>
                            </div>
                            <div className="field">
                                <label>Date</label>
                                <input className="input" type="date" value={takenOn} onChange={(e) => setTakenOn(e.target.value)} />
                            </div>
                        </div>

                        <div className="rd-round-row rd-round-head">
                            <span>Meter</span>
                            <span className="num">Previous</span>
                            <span>New value</span>
                            <span className="num">Consumption</span>
                        </div>

                        {groups.map((g) => (
                            <AccountRound
                                key={g.account.id}
                                group={g}
                                values={values}
                                latestBy={latestBy}
                                consumptionOf={consumptionOf}
                                onValue={(key, v) => setValues({ ...values, [key]: v })}
                            />
                        ))}

                        {groups.length === 0 && (
                            <p className="xd-note">No active meters. Add them on the Meters page.</p>
                        )}

                        <div className="rd-form-foot">
                            <span className="mt-hint">
                                Meters left empty are skipped. A day/night meter needs both registers.
                            </span>
                            {status && <span className="status-ok">{status}</span>}
                            <button type="submit" className="btn btn-primary">Save readings</button>
                        </div>
                    </form>

                    <ReadingsHistory
                        meters={meters}
                        accounts={accounts}
                        serviceTypes={serviceTypes}
                        selectedId={selectedId}
                        onSelect={(id) => setSearchParams({ meter: id })}
                        reloadKey={historyKey}
                        onChanged={loadAll}
                    />
                </>
            )}
        </div>
    );
}

function AccountRound({ group, values, latestBy, consumptionOf, onValue }) {
    const { account, type, meters } = group;
    const unit = type ? type.unit : "";

    return (
        <div className="rd-account">
            <div className="rd-account-title">
                {type ? type.name : account.service} <span className="mono row-meta">· {account.number}</span>
            </div>

            {meters.flatMap((m) =>
                zonesOf(m).map((z) => {
                    const key = registerKey(m.id, z);
                    const prev = latestBy[key];
                    const c = consumptionOf(key);
                    return (
                        <div key={key} className="rd-round-row">
                            <span className="rd-register">
                                <span className={m.name ? "" : "mono"}>{meterTitle(m)}</span>
                                {zoneLabel(z) && <span className="mt-zone-badge dual">{zoneLabel(z).toLowerCase()}</span>}
                            </span>
                            <span className="num mono row-meta">
                                {prev ? `${formatReading(prev.value)} · ${formatISODate(prev.taken_on)}` : "—"}
                            </span>
                            <input
                                className="input mono"
                                type="number"
                                step="any"
                                min="0"
                                placeholder={unit}
                                value={values[key] ?? ""}
                                onChange={(e) => onValue(key, e.target.value)}
                            />
                            <span className={c === null ? "num mono mt-delta none" : c < 0 ? "num mono rd-negative" : "num mono mt-delta"}>
                                {c === null ? "—" : `${c < 0 ? "−" : "+"}${formatReading(Math.abs(c))}`}
                            </span>
                        </div>
                    );
                })
            )}
        </div>
    );
}

export default ReadingsPage;
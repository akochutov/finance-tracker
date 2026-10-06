import { useState, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import {
    getServiceTypes, getAddresses, getUtilityAccounts, getMeters, getLatestReadings,
    createReadings, setZoneUsage,
} from "../api/client";
import { todayISO, formatISODate } from "./expenseUtils";
import { isDayNight, billingMonthKey, monthLabel, formatReading, sameAmount } from "./meterUtils";
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
    const [splits, setSplits] = useState({});

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
    const latestBy = Object.fromEntries(latest.map((r) => [r.meter_id, r]));
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

    function consumptionOf(meter) {
        const raw = values[meter.id];
        const prev = latestBy[meter.id];
        if (raw === undefined || raw === "" || !prev) return null;
        return Number(raw) - Number(prev.value);
    }

    async function save(e) {
        e.preventDefault();
        setError(null);
        setStatus(null);

        const readings = Object.entries(values)
            .filter(([, v]) => v !== "")
            .map(([meterId, v]) => ({ meter_id: meterId, value: v }));
        const filledSplits = Object.entries(splits).filter(([, s]) => s.day !== "" && s.night !== "" && s.day !== undefined && s.night !== undefined);

        if (readings.length === 0 && filledSplits.length === 0) {
            setError("Nothing to save: enter at least one value.");
            return;
        }

        try {
            if (readings.length > 0) {
                await createReadings({ taken_on: takenOn, readings });
            }
            for (const [accountId, s] of filledSplits) {
                await setZoneUsage(accountId, month, { day: s.day, night: s.night });
            }
            setValues({});
            setSplits({});
            setStatus(
                `Saved ${readings.length} ${readings.length === 1 ? "reading" : "readings"}` +
                (filledSplits.length ? ` and the provider split for ${monthLabel(month)}` : "") + "."
            );
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
                                onValue={(meterId, v) => setValues({ ...values, [meterId]: v })}
                                split={splits[g.account.id] || { day: "", night: "" }}
                                onSplit={(s) => setSplits({ ...splits, [g.account.id]: s })}
                                month={month}
                            />
                        ))}

                        {groups.length === 0 && (
                            <p className="xd-note">No active meters. Add them on the Meters page.</p>
                        )}

                        <div className="rd-form-foot">
                            <span className="mt-hint">
                                Meters left empty are skipped. A provider split can be added later, when the bill arrives.
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

function AccountRound({ group, values, latestBy, consumptionOf, onValue, split, onSplit, month }) {
    const { account, type, meters } = group;
    const unit = type ? type.unit : "";
    const dayNight = isDayNight(account);

    const consumptions = meters.map(consumptionOf);
    const total = consumptions.every((c) => c !== null) ? consumptions.reduce((a, b) => a + b, 0) : null;
    const splitFilled = split.day !== "" && split.night !== "";
    const splitSum = splitFilled ? Number(split.day) + Number(split.night) : null;

    let check = null;
    if (splitFilled && total !== null) {
        const diff = splitSum - total;
        check = sameAmount(splitSum, total)
            ? <span className="rd-check ok">✓ {formatReading(total)} — matches the meter</span>
            : <span className="rd-check bad">{diff > 0 ? "+" : "−"}{formatReading(Math.abs(diff))} {unit} against the meter</span>;
    }

    return (
        <div className="rd-account">
            <div className="rd-account-title">
                {type ? type.name : account.service} <span className="mono row-meta">· {account.number}</span>
                {dayNight && <span className="mt-zone-badge dual">day/night</span>}
            </div>

            {meters.map((m) => {
                const prev = latestBy[m.id];
                const c = consumptionOf(m);
                return (
                    <div key={m.id} className="rd-round-row">
                        <span className="mono">{m.serial}</span>
                        <span className="num mono row-meta">
                            {prev ? `${formatReading(prev.value)} · ${formatISODate(prev.taken_on)}` : "—"}
                        </span>
                        <input
                            className="input mono"
                            type="number"
                            step="any"
                            min="0"
                            placeholder={unit}
                            value={values[m.id] ?? ""}
                            onChange={(e) => onValue(m.id, e.target.value)}
                        />
                        <span className={c === null ? "num mono mt-delta none" : c < 0 ? "num mono rd-negative" : "num mono mt-delta"}>
                            {c === null ? "—" : `${c < 0 ? "−" : "+"}${formatReading(Math.abs(c))}`}
                        </span>
                    </div>
                );
            })}

            {dayNight && (
                <div className="rd-split">
                    <span className="mt-info-label">Provider split · {monthLabel(month)}</span>
                    <label className="rd-split-field">
                        day
                        <input
                            className="input mono"
                            type="number"
                            step="any"
                            min="0"
                            value={split.day}
                            onChange={(e) => onSplit({ ...split, day: e.target.value })}
                        />
                    </label>
                    <label className="rd-split-field">
                        night
                        <input
                            className="input mono"
                            type="number"
                            step="any"
                            min="0"
                            value={split.night}
                            onChange={(e) => onSplit({ ...split, night: e.target.value })}
                        />
                    </label>
                    {check}
                </div>
            )}
        </div>
    );
}

export default ReadingsPage;
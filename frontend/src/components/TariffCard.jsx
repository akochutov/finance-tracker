import { useState } from "react";
import { formatISODate, todayISO } from "./expenseUtils";
import TierEditor from "./TierEditor";
import {
    MODE_LABELS, isFlat, formatPrice, tierRanges, tiersSummary, tiersFromTariff, tiersToRequest,
} from "./tariffUtils";

function TariffCard({ service, serviceCode, zone, tariffs, currencies, actions }) {
    const today = todayISO();
    const current = tariffs.find((t) => t.valid_from.slice(0, 10) <= today) || null;
    const unit = service ? service.unit : "";
    const title = `${service ? service.name : serviceCode} · ${zone}`;

    return (
        <section className="xd-panel">
            <div className="tf-head">
                <span className="tf-title">{title}</span>
                <span className="row-meta">{tariffs.length} {tariffs.length === 1 ? "price" : "prices"}</span>
            </div>

            {current ? (
                <div className="tf-current">
                    <div className="tf-current-meta">
                        <span className="mt-info-label">In force</span>
                        <span className="row-meta">
                            since {formatISODate(current.valid_from)} · {current.currency} per {unit}
                        </span>
                        {!isFlat(current) && <span className="mt-pill accent">{MODE_LABELS[current.tier_mode]}</span>}
                    </div>
                    <div className="tf-tiers">
                        {tierRanges(current, unit).map((r) => (
                            <div key={r.label} className="tf-tier-line">
                                <span>{r.label}</span>
                                <span className="mono tf-tier-price">{formatPrice(r.price)} {current.currency}</span>
                            </div>
                        ))}
                    </div>
                </div>
            ) : (
                <p className="xd-note">No price in force yet.</p>
            )}

            <div className="tf-history">
                {tariffs.map((t) => (
                    <TariffRow
                        key={t.id}
                        tariff={t}
                        unit={unit}
                        isCurrent={current !== null && t.id === current.id}
                        isUpcoming={t.valid_from.slice(0, 10) > today}
                        currencies={currencies}
                        actions={actions}
                    />
                ))}
            </div>
        </section>
    );
}

function TariffRow({ tariff, unit, isCurrent, isUpcoming, currencies, actions }) {
    const [editing, setEditing] = useState(false);
    const [validFrom, setValidFrom] = useState(tariff.valid_from.slice(0, 10));
    const [currency, setCurrency] = useState(tariff.currency);
    const [mode, setMode] = useState(tariff.tier_mode);
    const [tiers, setTiers] = useState(tiersFromTariff(tariff));

    function startEdit() {
        setValidFrom(tariff.valid_from.slice(0, 10));
        setCurrency(tariff.currency);
        setMode(tariff.tier_mode);
        setTiers(tiersFromTariff(tariff));
        setEditing(true);
    }

    async function save() {
        const ok = await actions.update(tariff.id, {
            valid_from: validFrom,
            currency,
            tier_mode: mode,
            tiers: tiersToRequest(tiers),
        });
        if (ok) setEditing(false);
    }

    async function remove() {
        if (window.confirm(`Delete the price from ${formatISODate(tariff.valid_from)}?`)) {
            await actions.remove(tariff.id);
        }
    }

    if (editing) {
        return (
            <div className="tf-edit">
                <div className="tf-edit-fields">
                    <div className="field">
                        <label>Valid from</label>
                        <input className="input" type="date" value={validFrom} onChange={(e) => setValidFrom(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Currency</label>
                        <select className="input" value={currency} onChange={(e) => setCurrency(e.target.value)}>
                            {currencies.map((c) => (
                                <option key={c.code} value={c.code}>{c.code}</option>
                            ))}
                        </select>
                    </div>
                </div>
                <TierEditor
                    tiers={tiers}
                    onChange={setTiers}
                    mode={mode}
                    onModeChange={setMode}
                    unit={unit}
                    currency={currency}
                />
                <div className="tf-edit-actions">
                    <button type="button" className="btn btn-primary btn-sm" onClick={save}>Save</button>
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => setEditing(false)}>Cancel</button>
                </div>
            </div>
        );
    }

    const classes = ["tf-row"];
    if (isCurrent) classes.push("current");
    if (!isCurrent && !isUpcoming) classes.push("past");

    return (
        <div className={classes.join(" ")}>
            <span>{formatISODate(tariff.valid_from)}</span>
            <span className="mono num">{tiersSummary(tariff)} {tariff.currency}</span>
            <span>
                {isCurrent && <span className="mt-pill accent">current</span>}
                {isUpcoming && <span className="mt-pill">upcoming</span>}
            </span>
            <span className="mt-row-actions">
                <button type="button" className="mt-link" onClick={startEdit}>edit</button>
                <button type="button" className="mt-link mt-link-danger" onClick={remove}>delete</button>
            </span>
        </div>
    );
}

export default TariffCard;
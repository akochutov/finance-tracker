import { useState } from "react";
import { formatISODate, todayISO } from "./expenseUtils";

function formatPrice(value) {
    return Number(value).toLocaleString(undefined, { maximumFractionDigits: 4 });
}

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
                    <span className="tf-current-price mono">{formatPrice(current.price)}</span>
                    <span className="row-meta">
                        {current.currency} / {unit} · since {formatISODate(current.valid_from)}
                    </span>
                </div>
            ) : (
                <p className="xd-note">No price in force yet.</p>
            )}

            <div className="tf-history">
                {tariffs.map((t) => (
                    <TariffRow
                        key={t.id}
                        tariff={t}
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

function TariffRow({ tariff, isCurrent, isUpcoming, currencies, actions }) {
    const [editing, setEditing] = useState(false);
    const [validFrom, setValidFrom] = useState(tariff.valid_from.slice(0, 10));
    const [price, setPrice] = useState(String(tariff.price));
    const [currency, setCurrency] = useState(tariff.currency);

    function startEdit() {
        setValidFrom(tariff.valid_from.slice(0, 10));
        setPrice(String(tariff.price));
        setCurrency(tariff.currency);
        setEditing(true);
    }

    async function save() {
        const ok = await actions.update(tariff.id, {
            valid_from: validFrom,
            price: price === "" ? null : price,
            currency,
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
            <div className="tf-row tf-row-editing">
                <input className="input" type="date" value={validFrom} onChange={(e) => setValidFrom(e.target.value)} />
                <input className="input mono" type="number" step="any" min="0" value={price} onChange={(e) => setPrice(e.target.value)} />
                <select className="input" value={currency} onChange={(e) => setCurrency(e.target.value)}>
                    {currencies.map((c) => (
                        <option key={c.code} value={c.code}>{c.code}</option>
                    ))}
                </select>
                <span className="mt-row-actions">
                    <button type="button" className="mt-link" onClick={save}>save</button>
                    <button type="button" className="mt-link" onClick={() => setEditing(false)}>cancel</button>
                </span>
            </div>
        );
    }

    const classes = ["tf-row"];
    if (isCurrent) classes.push("current");
    if (!isCurrent && !isUpcoming) classes.push("past");

    return (
        <div className={classes.join(" ")}>
            <span>{formatISODate(tariff.valid_from)}</span>
            <span className="mono num">{formatPrice(tariff.price)} {tariff.currency}</span>
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
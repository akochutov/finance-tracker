import { useState, useEffect } from "react";
import { getTariffs, createTariff, updateTariff, deleteTariff, getServiceTypes, getCurrencies } from "../api/client";
import { todayISO } from "./expenseUtils";
import TariffCard from "./TariffCard";
import TierEditor from "./TierEditor";
import { ZONES, MODE_WHOLE, emptyTiers, tiersToRequest } from "./tariffUtils";

const ZONE_ORDER = { single: 0, day: 1, night: 2 };

function TariffsPage() {
    const [tariffs, setTariffs] = useState([]);
    const [serviceTypes, setServiceTypes] = useState([]);
    const [currencies, setCurrencies] = useState([]);
    const [loaded, setLoaded] = useState(false);
    const [error, setError] = useState(null);

    const [service, setService] = useState("electricity");
    const [zone, setZone] = useState("single");
    const [validFrom, setValidFrom] = useState(todayISO());
    const [currency, setCurrency] = useState("AMD");
    const [tiers, setTiers] = useState(emptyTiers());
    const [mode, setMode] = useState(MODE_WHOLE);

    async function loadTariffs() {
        setTariffs(await getTariffs());
    }

    useEffect(() => {
        async function loadAll() {
            try {
                const [types, curs, list] = await Promise.all([getServiceTypes(), getCurrencies(), getTariffs()]);
                setServiceTypes(types);
                setCurrencies(curs.filter((c) => c.is_active));
                setTariffs(list);
                setLoaded(true);
            } catch (err) {
                setError(err.message);
            }
        }
        loadAll();
    }, []);

    async function run(action) {
        setError(null);
        try {
            await action();
            await loadTariffs();
            return true;
        } catch (err) {
            setError(err.message);
            window.scrollTo({ top: 0, behavior: "smooth" });
            return false;
        }
    }

    async function handleCreate(e) {
        e.preventDefault();
        const ok = await run(() =>
            createTariff({
                service,
                zone,
                valid_from: validFrom,
                currency,
                tier_mode: mode,
                tiers: tiersToRequest(tiers),
            })
        );
        if (ok) {
            setTiers(emptyTiers());
            setMode(MODE_WHOLE);
        }
    }

    const actions = {
        update: (id, fields) => run(() => updateTariff(id, fields)),
        remove: (id) => run(() => deleteTariff(id)),
    };

    const serviceOrder = Object.fromEntries(serviceTypes.map((t, i) => [t.code, i]));
    const groups = new Map();
    for (const t of tariffs) {
        const key = `${t.service}|${t.zone}`;
        if (!groups.has(key)) groups.set(key, { service: t.service, zone: t.zone, tariffs: [] });
        groups.get(key).tariffs.push(t);
    }
    const cards = [...groups.values()].sort(
        (a, b) => (serviceOrder[a.service] ?? 99) - (serviceOrder[b.service] ?? 99) || ZONE_ORDER[a.zone] - ZONE_ORDER[b.zone]
    );

    return (
        <div className="xd-page">
            <div className="page-header">
                <h1>Tariffs</h1>
                <p>Prices per unit, flat or by monthly volume.</p>
            </div>

            {error && <div className="error">{error}</div>}
            {!loaded && !error && <div className="loading">Loading…</div>}

            {loaded && (
                <>
                    <form className="card" onSubmit={handleCreate}>
                        <h5 className="form-title">Add tariff</h5>
                        <div className="form-grid">
                            <div className="field">
                                <label>Service</label>
                                <select className="input" value={service} onChange={(e) => setService(e.target.value)}>
                                    {serviceTypes.map((t) => (
                                        <option key={t.code} value={t.code}>{t.name}</option>
                                    ))}
                                </select>
                            </div>
                            <div className="field">
                                <label>Zone</label>
                                <select className="input" value={zone} onChange={(e) => setZone(e.target.value)}>
                                    {ZONES.map((z) => (
                                        <option key={z} value={z}>{z}</option>
                                    ))}
                                </select>
                            </div>
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
                            unit={serviceTypes.find((t) => t.code === service)?.unit}
                            currency={currency}
                        />
                        <div className="form-actions">
                            <button type="submit" className="btn btn-primary">Add</button>
                        </div>
                    </form>

                    {cards.length === 0 ? (
                        <p className="xd-note">No tariffs yet. Add the current price of each service from your bills.</p>
                    ) : (
                        <div className="tf-grid">
                            {cards.map((c) => (
                                <TariffCard
                                    key={`${c.service}|${c.zone}`}
                                    service={serviceTypes.find((t) => t.code === c.service)}
                                    serviceCode={c.service}
                                    zone={c.zone}
                                    tariffs={c.tariffs}
                                    currencies={currencies}
                                    actions={actions}
                                />
                            ))}
                        </div>
                    )}
                </>
            )}
        </div>
    );
}

export default TariffsPage;
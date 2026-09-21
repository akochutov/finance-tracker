import { useState, useEffect } from "react";
import { getSettings, updateSettings, getCurrencies } from "../api/client";
import RateSourceSettings from "./RateSourceSettings";

function SettingsPage() {
    const [baseCurrency, setBaseCurrency] = useState("");
    const [fiatCurrencies, setFiatCurrencies] = useState([]);
    const [status, setStatus] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        Promise.all([getSettings(), getCurrencies()])
            .then(([settings, currencies]) => {
                setBaseCurrency(settings.base_currency);
                setFiatCurrencies(
                    currencies.filter((c) => c.kind === "fiat" && c.is_active)
                );
            })
            .catch((err) => setError(err.message));
    }, []);

    async function handleChange(event) {
        const code = event.target.value;
        const previous = baseCurrency;

        setBaseCurrency(code);
        setStatus(null);
        try {
            await updateSettings(code);
            setStatus({ type: "ok", text: `Base currency set to ${code}` });
        } catch (err) {
            setBaseCurrency(previous);
            setStatus({ type: "error", text: err.message });
        }
    }

    if (error) {
        return <div className="page"> Error loading settings: {error}</div>;
    }

    return (
        <div className="page">
            <h1>Settings</h1>

            <div className="settings-field">
                <label htmlFor="base-currency">Base currency</label>
                <select
                    id="base-currency"
                    value={baseCurrency}
                    onChange={handleChange}
                >
                    {fiatCurrencies.map((c) => (
                        <option key={c.code} value={c.code}>
                            {c.code} - {c.name}
                        </option>
                    ))}
                </select>
                <p className="settings-hint">
                    Dashboards convert to this currency. Existing records are never changed.
                </p>
            </div>

            {status && (
                <p className={status.type === "ok" ? "status-ok" : "status-error"}>
                    {status.text}
                </p>
            )}
            <RateSourceSettings />
        </div>
    );
}

export default SettingsPage;
import { useState, useEffect } from "react";
import { getSettings, updateSettings, updateExpenseBaseCurrency, getCurrencies } from "../api/client";
import RateSourceSettings from "./RateSourceSettings";

function SettingsPage() {
    const [baseCurrency, setBaseCurrency] = useState("");
    const [expenseCurrency, setExpenseCurrency] = useState("");
    const [fiatCurrencies, setFiatCurrencies] = useState([]);
    const [status, setStatus] = useState(null);
    const [error, setError] = useState(null);

    useEffect(() => {
        Promise.all([getSettings(), getCurrencies()])
            .then(([settings, currencies]) => {
                setBaseCurrency(settings.base_currency);
                setExpenseCurrency(settings.expense_base_currency);
                setFiatCurrencies(
                    currencies.filter((c) => c.kind === "fiat" && c.is_active)
                );
            })
            .catch((err) => setError(err.message));
    }, []);

    async function changeCurrency(code, previous, setValue, save, label) {
        setValue(code);
        setStatus(null);
        try {
            await save(code);
            setStatus({ type: "ok", text: `${label} set to ${code}` });
        } catch (err) {
            setValue(previous);
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
                <label htmlFor="base-currency">Income dashboard currency</label>
                <select
                    id="base-currency"
                    value={baseCurrency}
                    onChange={(e) => changeCurrency(e.target.value, baseCurrency, setBaseCurrency, updateSettings, "Income dashboard currency")}
                >
                    {fiatCurrencies.map((c) => (
                        <option key={c.code} value={c.code}>
                            {c.code} - {c.name}
                        </option>
                    ))}
                </select>
                <p className="settings-hint">
                    The income dashboard converts to this currency. Existing records are never changed.
                </p>
            </div>

            <div className="settings-field">
                <label htmlFor="expense-currency">Expenses dashboard currency</label>
                <select
                    id="expense-currency"
                    value={expenseCurrency}
                    onChange={(e) => changeCurrency(e.target.value, expenseCurrency, setExpenseCurrency, updateExpenseBaseCurrency, "Expenses dashboard currency")}
                >
                    {fiatCurrencies.map((c) => (
                        <option key={c.code} value={c.code}>
                            {c.code} - {c.name}
                        </option>
                    ))}
                </select>
                <p className="settings-hint">
                    The expenses dashboard converts to this currency. Until you choose one, it follows the income currency.
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
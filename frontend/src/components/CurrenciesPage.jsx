import { useState, useEffect } from "react";
import { deactivateCurrency, getCurrencies, updateCurrency } from "../api/client";
import CurrenciesList from "./CurrenciesList";
import CurrencyForm from "./CurrencyForm";

function CurrenciesPage() {
    const [currencies, setCurrencies] = useState([]);
    const [error, setError] = useState(null);

    async function loadCurrencies() {
        try {
            const data = await getCurrencies();
            setCurrencies(data);
        } catch (err) {
            setError(err.message);
        }
    }

    async function handleDeactivate(code) {
        try {
            await deactivateCurrency(code);
            await loadCurrencies();
        } catch (err) {
            setError(err.message);
        }
    }

    async function handleSave(code, fields) {
        try {
            await updateCurrency(code, fields);
            await loadCurrencies();
        } catch (err) {
            setError(err.message);
        }
    }

    useEffect(() => {
        loadCurrencies();
    }, []);

    return (
        <div>
            <h2>Currencies</h2>
            {error && <div style={{ color: "red" }}>{error}</div>}
            <CurrencyForm onCreated={loadCurrencies} />
            <CurrenciesList 
                currencies={currencies} 
                onSave={handleSave}
                onDeactivate={handleDeactivate} 
            />
        </div>
    );
}

export default CurrenciesPage;
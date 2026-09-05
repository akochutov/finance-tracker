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
            <div className="page-header">
                <h1>Currencies</h1>
                <p>Units of account used across incomes and expenses.</p>
            </div>
            {error && <div className="error">{error}</div>}
            <CurrencyForm onCreated={loadCurrencies} />
            <div className="list-header">
                <h5>All currencies</h5>
                <span className="row-meta">{currencies.length}</span>
            </div>
            <CurrenciesList
                currencies={currencies}
                onSave={handleSave}
                onDeactivate={handleDeactivate}
            />
        </div>
    );
}

export default CurrenciesPage;
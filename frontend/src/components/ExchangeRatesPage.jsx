import { useState, useEffect, useCallback } from "react";
import { getExchangeRates, getCurrencies } from "../api/client";
import ExchangeRateForm from "./ExchangeRateForm";
import ExchangeRateRow from "./ExchangeRateRow";

function ExchangeRatesPage() {
    const [rates, setRates] = useState([]);
    const [currencies, setCurrencies] = useState([]);
    const [filter, setFilter] = useState("");
    const [error, setError] = useState(null);

    const loadRates = useCallback(() => {
        getExchangeRates(filter)
            .then(setRates)
            .catch((err) => setError(err.message));
    }, [filter]);

    useEffect(() => {
        getCurrencies()
            .then(setCurrencies)
            .catch((err) => setError(err.message));
    }, []);

    useEffect(() => {
        loadRates();
    }, [loadRates]);

    const filterOptions = currencies.filter(
        (c) => c.code !== "USD" && c.is_active
    );

    return (
        <>
            <div className="page-header">
                <h1>Exchange Rates</h1>
                <p>Rates are stored as the price of one unit in USD.</p>
            </div>

            {error && <div className="error">{error}</div>}

            <ExchangeRateForm currencies={currencies} onCreated={loadRates} />

            <div className="filter-bar">
                <div className="field">
                    <label htmlFor="er-filter">Filter by currency</label>
                    <select
                        id="er-filter"
                        value={filter}
                        onChange={(e) => setFilter(e.target.value)}
                    >
                        <option value="">All currencies</option>
                        {filterOptions.map((c) => (
                            <option key={c.code} value={c.code}>
                                {c.code}
                            </option>
                        ))}
                    </select>
                </div>
            </div>

            <ul className="list">
                {rates.length === 0 ? (
                    <li className="loading">No rates yet.</li>
                ) : (
                    rates.map((r) => (
                        <ExchangeRateRow
                            key={`${r.currency}-${r.source}-${r.rate_at}`}
                            rate={r}
                        />
                    ))
                )}
            </ul>
        </>
    );
}

export default ExchangeRatesPage;
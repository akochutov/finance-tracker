import { useState, useEffect } from "react";
import { getCurrencies } from "../api/client";

function CurrenciesList() {
    const [currencies, setCurrencies] = useState([]);
    const [error, setError] = useState(null);

    useEffect(() => {
        getCurrencies()
            .then((data) => setCurrencies(data))
            .catch((err) => setError(err.message));
    }, [])

    if (error) {
        return <div>Error loading currencies: {error}</div>;
    }

    return (
        <div>
            <h2>Currencies</h2>
            <ul>
                {currencies.map((c) => (
                    <li key={c.code}>
                        {c.code} - {c.name} ({c.kind}, {c.decimal_places} decimals)
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default CurrenciesList;
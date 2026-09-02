import { useState, useEffect } from "react";
import { getIncomes } from "../api/client";

function formatDate(isoString) {
    return new Date(isoString).toLocaleDateString();
}

function IncomesList() {
    const [incomes, setIncomes] = useState([]);
    const [error, setError] = useState(null);

    useEffect(() => {
        getIncomes()
            .then((data) => setIncomes(data))
            .catch((err) => setError(err.message));
    }, [])

    if (error) {
        return <div>Error loading incomes: {error}</div>;
    }

    return (
        <div>
            <h2>Incomes</h2>
            <ul>
                {incomes.map((inc) => (
                    <li key={inc.id}>
                        {formatDate(inc.occurred_at)}: {inc.amount} {inc.currency} ({inc.payment_type})
                        {inc.note ? ` - ${inc.note}` : ""}
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default IncomesList;
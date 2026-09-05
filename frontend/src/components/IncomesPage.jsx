import { useState, useEffect } from "react";
import { getIncomes, getCompanies } from "../api/client";
import IncomeForm from "./IncomeForm";

function formatDate(isoString) {
    return new Date(isoString).toLocaleDateString();
}

function IncomesPage() {
    const [incomes, setIncomes] = useState([]);
    const [companiesById, setCompaniesById] = useState({});
    const [error, setError] = useState(null);

    async function loadData() {
        try {
            const companies = await getCompanies();
            const lookup = {};
            for (const c of companies) {
                lookup[c.id] = c.name;
            }
            setCompaniesById(lookup);

            const data = await getIncomes();
            setIncomes(data);
        } catch (err) {
            setError(err.message);
        }
    }

    useEffect(() => {
        loadData();
    }, []);

    if (error) {
        return <div>Error: {error}</div>
    }

    return (
        <div>
            <h2>Incomes</h2>
            <IncomeForm onCreated={loadData} /> 
            <ul>
                {incomes.map((inc) => (
                    <li key={inc.id}>
                        {formatDate(inc.occurred_at)}: {inc.amount} {inc.currency} —{" "}
                        from {companiesById[inc.payer_id] || inc.payer_id}{" "}
                        to {companiesById[inc.beneficiary_id] || inc.beneficiary_id}{" "}
                        ({inc.payment_type})
                        {inc.note ? ` — ${inc.note}` : ""}
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default IncomesPage;
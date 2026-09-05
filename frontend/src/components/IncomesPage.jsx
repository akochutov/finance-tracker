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
        return <div className="error">{error}</div>;
    }

    return (
        <div>
            <div className="page-header">
                <h1>Incomes</h1>
                <p>Recorded payments between companies.</p>
            </div>
            <IncomeForm onCreated={loadData} />
            <div className="list-header">
                <h5>All incomes</h5>
                <span className="row-meta">{incomes.length}</span>
            </div>
            <ul className="list">
                {incomes.map((inc) => (
                    <li key={inc.id} className="row">
                        <span className="row-meta">{formatDate(inc.occurred_at)}</span>
                        <span className="row-key mono">{inc.amount} {inc.currency}</span>
                        <span>
                            {companiesById[inc.payer_id] || inc.payer_id}
                            {" → "}
                            {companiesById[inc.beneficiary_id] || inc.beneficiary_id}
                        </span>
                        <span className="badge">{inc.payment_type}</span>
                        {inc.note && <span className="row-meta">{inc.note}</span>}
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default IncomesPage;
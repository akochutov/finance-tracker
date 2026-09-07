import { useState, useEffect } from "react";
import { getIncomes, getCompanies } from "../api/client";
import IncomeForm from "./IncomeForm";
import IncomeRow from "./IncomeRow";

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
                    <IncomeRow key={inc.id} income={inc} companiesById={companiesById} />
                ))}
            </ul>
        </div>
    );
}

export default IncomesPage;
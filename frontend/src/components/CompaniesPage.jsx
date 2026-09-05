import { useState, useEffect } from "react";
import { getCompanies, updateCompany, deactivateCompany } from "../api/client";
import CompaniesList from "./CompaniesList";
import CompanyForm from "./CompanyForm";

function CompaniesPage() {
    const [companies, setCompanies] = useState([]);
    const [error, setError] = useState(null);

    async function loadCompanies() {
        try {
            const data = await getCompanies();
            setCompanies(data);
        } catch (err) {
            setError(err.message);
        }
    }

    async function handleSave(id, fields) {
        try {
            await updateCompany(id, fields);
            await loadCompanies();
        } catch (err) {
            setError(err.message);
        }
    }

    async function handleDeactivate(id) {
        try {
            await deactivateCompany(id);
            await loadCompanies();
        } catch (err) {
            setError(err.message);
        }
    }

    useEffect(() => {
        loadCompanies();
    }, []);

    return (
        <div>
            <div className="page-header">
                <h1>Companies</h1>
                <p>Counterparties: payers, beneficiaries and their requisites.</p>
            </div>
            {error && <div className="error">{error}</div>}
            <CompanyForm onCreated={loadCompanies} />
            <div className="list-header">
                <h5>All companies</h5>
                <span className="row-meta">{companies.length}</span>
            </div>
            <CompaniesList
                companies={companies}
                onSave={handleSave}
                onDeactivate={handleDeactivate}
            />
        </div>
    );
}

export default CompaniesPage;
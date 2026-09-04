import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { getCompany, getBankRequisites, getCryptoRequisites } from "../api/client";
import BankRequisiteForm from "./BankRequisiteForm";
import CryptoRequisiteForm from "./CryptoRequisiteForm";

function CompanyPage() {
    const { id } = useParams();
    const [company, setCompany] = useState(null);
    const [bankRequisites, setBankRequisites] = useState([]);
    const [cryptoRequisites, setCryptoRequisites] = useState([]);
    const [error, setError] = useState(null);

    async function loadData() {
        try {
            const c = await getCompany(id);
            setCompany(c);

            const bank = await getBankRequisites(id);
            setBankRequisites(bank);

            const crypto = await getCryptoRequisites(id);
            setCryptoRequisites(crypto);
        } catch (err) {
            setError(err.message);
        }
    }

    useEffect(() => {
        loadData();
    }, [id]);

    if (error) {
        return <div>Error: {error}</div>;
    }

    if (!company) {
        return <div>Loading...</div>
    }

    return (
        <div>
            <Link to="/companies">← Back to companies</Link>
            <h2>{company.name}</h2>
            <p>Tax ID: {company.tax_id || "-"}</p>
            <p>Address: {company.address || "-"}</p>
            <p>Note: {company.note || "-"}</p>

            <h3>Bank requisites</h3>
            <BankRequisiteForm companyId={id} onCreated={loadData} />
            <ul>
                {bankRequisites.map((r) => (
                    <li key={r.id}>
                        {r.bank_name} - {r.account_number} ({r.beneficiary_name})
                        {r.valid_to && ` [closed ${new Date(r.valid_to).toLocaleDateString()}]`}
                    </li>
                ))}
            </ul>

            <h3>Crypto requisites</h3>
            <CryptoRequisiteForm companyId={id} onCreated={loadData} />
            <ul>
                {cryptoRequisites.map((r) => (
                    <li key={r.id}>
                        {r.network} - {r.wallet_address}
                        {r.valid_to && ` [closed ${new Date(r.valid_to).toLocaleDateString()}]`}
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default CompanyPage;
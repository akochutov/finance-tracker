import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { getCompany, getBankRequisites, getCryptoRequisites,
    closeBankRequisite, closeCryptoRequisite } from "../api/client";
import RequisiteRow from "./RequisiteRow";
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

    async function handleCloseBank(requisiteId, validTo) {
        await closeBankRequisite(id, requisiteId, validTo);
        await loadData();        
    }

    async function handleCloseCrypto(requisiteId, validTo) {
        await closeCryptoRequisite(id, requisiteId, validTo);
        await loadData();        
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
                    <RequisiteRow 
                        key={r.id}
                        requisite={r}
                        label={`${r.bank_name} - ${r.account_number} (${r.beneficiary_name})`}
                        onClose={handleCloseBank}
                    />
                ))}
            </ul>

            <h3>Crypto requisites</h3>
            <CryptoRequisiteForm companyId={id} onCreated={loadData} />
            <ul>
                {cryptoRequisites.map((r) => (
                    <RequisiteRow
                        key={r.id}
                        requisite={r}
                        label={`${r.network} - ${r.wallet_address}`}
                        onClose={handleCloseCrypto}
                    />
                ))}
            </ul>
        </div>
    );
}

export default CompanyPage;
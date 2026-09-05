import { useState, useEffect } from "react";
import { createIncome, getCompanies, getCurrencies, getBankRequisites, getCryptoRequisites } from "../api/client";

function IncomeForm({ onCreated }) {
    const [companies, setCompanies] = useState([]);
    const [currencies, setCurrencies] = useState([]);

    const [payerId, setPayerId] = useState("");
    const [beneficiaryId, setBeneficiaryId] = useState("");
    const [amount,  setAmount] = useState("");
    const [currency, setCurrency] = useState("");
    const [occurredAt, setOccurredAt] = useState("");
    const [paymentType, setPaymentType] = useState("bank");
    const [note, setNote] = useState("");
    const [error, setError] = useState(null);

    const [payerRequisites, setPayerRequisites] = useState([]);
    const [beneficiaryRequisites, setBeneficiaryRequisites] = useState([]);
    const [payerRequisiteId, setPayerRequisiteId] = useState("");
    const [beneficiaryRequisiteId, setBeneficiaryRequisiteId] = useState("");

    useEffect(() => {
        async function loadOptions() {
            try {
                setCompanies(await getCompanies());
                setCurrencies(await getCurrencies());
            } catch (err) {
                setError(err.message);
            }
        }
        loadOptions();
    }, []);

    useEffect(() => {
        async function loadPayerRequisites() {
            if (!payerId) {
                setPayerRequisites([]);
                return;
            }
            try {
                const reqs = paymentType === "bank"
                    ? await getBankRequisites(payerId)
                    : await getCryptoRequisites(payerId);
                setPayerRequisites(reqs);
                setPayerRequisiteId("");
            } catch (err) {
                setError(err.message);
            }
        }
        loadPayerRequisites();
    }, [payerId, paymentType]);

    useEffect(() => {
        async function loadBeneficiaryRequisites() {
            if (!beneficiaryId) {
                setBeneficiaryRequisites([]);
                return;
            }
            try {
                const reqs = paymentType === "bank"
                    ? await getBankRequisites(beneficiaryId)
                    : await getCryptoRequisites(beneficiaryId);
                setBeneficiaryRequisites(reqs);
                setBeneficiaryRequisiteId("");
            } catch (err) {
                setError(err.message);
            }
        }
        loadBeneficiaryRequisites();
    }, [beneficiaryId, paymentType]);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await createIncome({
                payer_id: payerId,
                beneficiary_id: beneficiaryId,
                amount: amount,
                currency: currency,
                occurred_at: `${occurredAt}T00:00:00Z`,
                payment_type: paymentType,
                payer_requisite_id: payerRequisiteId,
                beneficiary_requisite_id: beneficiaryRequisiteId,
                note: note || null,
            });
            setPayerId("");
            setBeneficiaryId("");
            setAmount("");
            setCurrency("");
            setOccurredAt("");
            setPaymentType("bank");
            setNote("");
            setPayerRequisiteId("");
            setBeneficiaryRequisiteId("");
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form onSubmit={handleSubmit}>
            <h3>Add income</h3>
            {error && <div style={{ color: "red" }}>{error}</div>}
            
            <select value={payerId} onChange={(e) => setPayerId(e.target.value)}>
                <option value="">- Payer -</option>
                {companies.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                ))}
            </select>

            <select value={beneficiaryId} onChange={(e) => setBeneficiaryId(e.target.value)}>
                <option value="">- Beneficiary -</option>
                {companies.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                ))}
            </select>

            <select value={payerRequisiteId} onChange={(e) => setPayerRequisiteId(e.target.value)}>
                <option value="">- Payer requisite -</option>
                {payerRequisites.map((r) => (
                    <option key={r.id} value={r.id}>
                        {paymentType === "bank"
                            ? `${r.bank_name} - ${r.account_number}`
                            : `${r.network} - ${r.wallet_address}`}
                    </option>
                ))}
            </select>

            <select value={beneficiaryRequisiteId} onChange={(e) => setBeneficiaryRequisiteId(e.target.value)}>
                <option value="">- Beneficiary requisite -</option>
                {beneficiaryRequisites.map((r) => (
                    <option key={r.id} value={r.id}>
                        {paymentType === "bank"
                            ? `${r.bank_name} - ${r.account_number}`
                            : `${r.network} - ${r.wallet_address}`}
                    </option>
                ))}
            </select>

            <input 
                type="number"
                placeholder="Amount"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
            />

            <select value={currency} onChange={(e) => setCurrency(e.target.value)}>
                <option value="">- Currency -</option>
                {currencies.map((c) => (
                    <option key={c.code} value={c.code}>{c.code}</option>
                ))}
            </select>

            <input
                type="date"
                value={occurredAt}
                onChange={(e) => setOccurredAt(e.target.value)} 
            />

            <select value={paymentType} onChange={(e) => setPaymentType(e.target.value)}>
                <option value="bank">bank</option>
                <option value="crypto">crypto</option>
            </select>

            <input
                placeholder="Note"
                value={note}
                onChange={(e) => setNote(e.target.value)}
            />

            <button type="submit">Create</button>
        </form>
    );
}

export default IncomeForm;
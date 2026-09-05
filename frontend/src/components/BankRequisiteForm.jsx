import { useState } from "react";
import { createBankRequisite } from "../api/client";

function BankRequisiteForm({ companyId, onCreated }) {
    const [beneficiaryName, setBeneficiaryName] = useState("");
    const [accountNumber, setAccountNumber] = useState("");
    const [bankName, setBankName] = useState("");
    const [bankSwift, setBankSwift] = useState("");
    const [bankAddress, setBankAddress] = useState("");
    const [correspondentBankName, setCorrespondentBankName] = useState("");
    const [correspondentBankSwift, setCorrespondentBankSwift] = useState("");
    const [intermediaryBankName, setIntermediaryBankName] = useState("");
    const [intermediaryBankSwift, setIntermediaryBankSwift] = useState("");
    const [error, setError] = useState(null);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await createBankRequisite(companyId, {
                beneficiary_name: beneficiaryName,
                account_number: accountNumber,
                bank_name: bankName,
                bank_swift: bankSwift,
                bank_address: bankAddress || null,
                correspondent_bank_name: correspondentBankName || null,
                correspondent_bank_swift: correspondentBankSwift || null,
                intermediary_bank_name: intermediaryBankName || null,
                intermediary_bank_swift: intermediaryBankSwift || null,
            });
            setBeneficiaryName("");
            setAccountNumber("");
            setBankName("");
            setBankSwift("");
            setBankAddress("");
            setCorrespondentBankName("");
            setCorrespondentBankSwift("");
            setIntermediaryBankName("");
            setIntermediaryBankSwift("");
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form className="card" onSubmit={handleSubmit}>
            <h5 className="form-title">Add bank requisite</h5>
            {error && <div className="error">{error}</div>}
            <div className="form-grid">
                <div className="field">
                    <label>Beneficiary Name</label>
                    <input className="input" value={beneficiaryName} onChange={(e) => setBeneficiaryName(e.target.value)} />
                </div>
                <div className="field">
                    <label>Account Number</label>
                    <input className="input" value={accountNumber} onChange={(e) => setAccountNumber(e.target.value)} />
                </div>
                <div className="field">
                    <label>Bank Name</label>
                    <input className="input" value={bankName} onChange={(e) => setBankName(e.target.value)} />
                </div>
                <div className="field">
                    <label>Bank SWIFT</label>
                    <input className="input" value={bankSwift} onChange={(e) => setBankSwift(e.target.value)} />
                </div>
                <div className="field">
                    <label>Bank Address</label>
                    <input className="input" value={bankAddress} onChange={(e) => setBankAddress(e.target.value)} />
                </div>
                <div className="field">
                    <label>Correspondent Bank Name</label>
                    <input className="input" value={correspondentBankName} onChange={(e) => setCorrespondentBankName(e.target.value)} />
                </div>
                <div className="field">
                    <label>Correspondent Bank SWIFT</label>
                    <input className="input" value={correspondentBankSwift} onChange={(e) => setCorrespondentBankSwift(e.target.value)} />
                </div>
                <div className="field">
                    <label>Intermediary Bank Name</label>
                    <input className="input" value={intermediaryBankName} onChange={(e) => setIntermediaryBankName(e.target.value)} />
                </div>
                <div className="field">
                    <label>Intermediary Bank SWIFT</label>
                    <input className="input" value={intermediaryBankSwift} onChange={(e) => setIntermediaryBankSwift(e.target.value)} />
                </div>
            </div>
            <div className="form-actions">
                <button type="submit" className="btn btn-primary">Add</button>
            </div>
        </form>
    );
} 

export default BankRequisiteForm;
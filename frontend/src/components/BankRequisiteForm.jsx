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
        <form onSubmit={handleSubmit}>
            {error && <div style={{ color: "red" }}>{error}</div>}
            <input
                placeholder="Beneficiary Name"
                value={beneficiaryName}
                onChange={(e) => setBeneficiaryName(e.target.value)}
            />
            <input
                placeholder="Account Number"
                value={accountNumber}
                onChange={(e) => setAccountNumber(e.target.value)}
            />
            <input
                placeholder="Bank Name"
                value={bankName}
                onChange={(e) => setBankName(e.target.value)}
            />
            <input
                placeholder="Bank SWIFT"
                value={bankSwift}
                onChange={(e) => setBankSwift(e.target.value)}
            />
            <input
                placeholder="Bank Address"
                value={bankAddress}
                onChange={(e) => setBankAddress(e.target.value)}
            />
            <input
                placeholder="Correspondent Bank Name"
                value={correspondentBankName}
                onChange={(e) => setCorrespondentBankName(e.target.value)}
            />
            <input
                placeholder="Correspondent Bank SWIFT"
                value={correspondentBankSwift}
                onChange={(e) => setCorrespondentBankSwift(e.target.value)}
            />
            <input
                placeholder="Intermediary Bank Name"
                value={intermediaryBankName}
                onChange={(e) => setIntermediaryBankName(e.target.value)}
            />
            <input
                placeholder="Intermediary Bank SWIFT"
                value={intermediaryBankSwift}
                onChange={(e) => setIntermediaryBankSwift(e.target.value)}
            />
            <button type="submit">Add</button>
        </form>
    );
} 

export default BankRequisiteForm;
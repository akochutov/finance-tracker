import { useState } from "react";
import { createCompany } from "../api/client";

function CompanyForm({ onCreated }) {
    const [name, setName] = useState("");
    const [note, setNote] = useState("");
    const [taxId, setTaxId] = useState("");
    const [address, setAddress] = useState("");
    const [error, setError] = useState(null);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await createCompany({
                name: name,
                note: note || null,
                tax_id: taxId || null,
                address: address || null,
            });
            setName("");
            setNote("");
            setTaxId("");
            setAddress("");
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form onSubmit={handleSubmit}>
            <h3>Add company</h3>
            {error && <div style={{ color: "red" }}>{error}</div>}
            <input
                placeholder="Name"
                value={name}
                onChange={(e) => setName(e.target.value)}
            />
            <input
                placeholder="Tax ID"
                value={taxId}
                onChange={(e) => setTaxId(e.target.value)}
            />
            <input
                placeholder="Address"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
            />
            <input
                placeholder="Note"
                value={note}
                onChange={(e) => setNote(e.target.value)}
            />
            <button type="submit">Create</button>
        </form>
    );
}

export default CompanyForm;
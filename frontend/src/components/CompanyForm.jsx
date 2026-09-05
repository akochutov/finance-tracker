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
        <form className="card" onSubmit={handleSubmit}>
            <h5 className="form-title">Add company</h5>
            {error && <div className="error">{error}</div>}
            <div className="form-grid">
                <div className="field">
                    <label>Name</label>
                    <input className="input" placeholder="Nitka Inc." value={name} onChange={(e) => setName(e.target.value)} />
                </div>
                <div className="field">
                    <label>Tax ID</label>
                    <input className="input" placeholder="Optional" value={taxId} onChange={(e) => setTaxId(e.target.value)} />
                </div>
                <div className="field">
                    <label>Address</label>
                    <input className="input" placeholder="Optional" value={address} onChange={(e) => setAddress(e.target.value)} />
                </div>
                <div className="field">
                    <label>Note</label>
                    <input className="input" placeholder="Optional" value={note} onChange={(e) => setNote(e.target.value)} />
                </div>
            </div>
            <div className="form-actions">
                <button type="submit" className="btn btn-primary">Create</button>
            </div>
        </form>
    );
}

export default CompanyForm;
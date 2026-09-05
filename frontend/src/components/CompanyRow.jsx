import { useState } from "react";
import { Link } from "react-router-dom";

function CompanyRow({ company, onSave, onDeactivate }) {
    const [isEditing, setIsEditing] = useState(false);
    const [name, setName] = useState(company.name);
    const [note, setNote] = useState(company.note || "");
    const [taxId, setTaxId] = useState(company.tax_id || "");
    const [address, setAddress] = useState(company.address || "");

    function startEdit() {
        setName(company.name);
        setNote(company.note || "");
        setTaxId(company.tax_id || "");
        setAddress(company.address || "");
        setIsEditing(true);
    }

    async function save() {
        await onSave(company.id, {
            name: name,
            note: note || null,
            tax_id: taxId || null,
            address: address || null,
        });
        setIsEditing(false);
    }

    if (isEditing) {
        return (
            <li className="row-editing">
                <div className="form-grid">
                    <div className="field">
                        <label>Name</label>
                        <input className="input" value={name} onChange={(e) => setName(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Tax ID</label>
                        <input className="input" value={taxId} onChange={(e) => setTaxId(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Address</label>
                        <input className="input" value={address} onChange={(e) => setAddress(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Note</label>
                        <input className="input" value={note} onChange={(e) => setNote(e.target.value)} />
                    </div>
                </div>
                <div className="form-actions">
                    <button className="btn btn-primary btn-sm" onClick={save}>Save</button>
                    <button className="btn btn-secondary btn-sm" onClick={() => setIsEditing(false)}>Cancel</button>
                </div>
            </li>
        );
    }

    return (
        <li className={company.is_active ? "row" : "row row-inactive"}>
            <Link to={`/companies/${company.id}`} className="row-key">{company.name}</Link>
            <span className="row-meta">({company.tax_id || "-"}, {company.address || "-"}, {company.note || "-"})</span>
            <span className="row-spacer" />
            {company.is_active ? (
                <div className="row-actions">
                    <button className="btn btn-secondary btn-sm" onClick={startEdit}>Edit</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => onDeactivate(company.id)}>Deactivate</button>
                </div>
            ) : (
                <span className="badge badge-inactive">inactive</span>
            )}
        </li>
    );
}

export default CompanyRow;
import { useState } from "react";

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
            <li>
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
                <button onClick={save}>Save</button>
                <button onClick={() => setIsEditing(false)}>Cancel</button>
            </li>
        );
    }

    return (
        <li>
            {company.name} ({company.tax_id || "-"}, {company.address || "-"}, {company.note || "-"})
            {!company.is_active && " [inactive]"}
            {company.is_active && (
                <>
                    <button onClick={startEdit}>Edit</button>
                    <button onClick={() => onDeactivate(company.id)}>Deactivate</button>
                </>
            )}
        </li>
    );
}

export default CompanyRow;
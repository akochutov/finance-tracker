import { useState } from "react";

function CurrencyRow({ currency, onSave, onDeactivate }) {
    const [isEditing, setIsEditing] = useState(false);
    const [name, setName] = useState(currency.name);
    const [decimalPlaces, setDecimalPlaces] = useState(currency.decimal_places);

    function startEdit() {
        setName(currency.name);
        setDecimalPlaces(currency.decimal_places);
        setIsEditing(true);
    }

    async function save() {
        await onSave(currency.code, {
            name: name,
            decimal_places: Number(decimalPlaces),
        });
        setIsEditing(false);
    }

    if (isEditing) {
        return (
            <li className="row-editing">
                <div className="form-grid">
                    <div className="field">
                        <label>Code</label>
                        <div className="row-key">{currency.code}</div>
                    </div>
                    <div className="field">
                        <label>Name</label>
                        <input className="input" value={name} onChange={(e) => setName(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Decimal places</label>
                        <input
                            className="input"
                            type="number"
                            value={decimalPlaces}
                            onChange={(e) => setDecimalPlaces(e.target.value)}
                        />
                    </div>
                    <div className="row-actions">
                        <button className="btn btn-primary btn-sm" onClick={save}>Save</button>
                        <button className="btn btn-secondary btn-sm" onClick={() => setIsEditing(false)}>Cancel</button>
                    </div>
                </div>
            </li>
        );
    }

    return (
        <li className={currency.is_active ? "row" : "row row-inactive"}>
            <span className="row-key">{currency.code}</span>
            <span>{currency.name}</span>
            <span className="row-meta">({currency.kind}, {currency.decimal_places} decimals)</span>
            <span className="row-spacer" />
            {currency.is_active ? (
                <div className="row-actions">
                    <button className="btn btn-secondary btn-sm" onClick={startEdit}>Edit</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => onDeactivate(currency.code)}>Deactivate</button>
                </div>
            ) : (
                <span className="badge badge-inactive">inactive</span>
            )}
        </li>
    );
}

export default CurrencyRow;
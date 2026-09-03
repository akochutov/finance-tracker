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
            <li>
                {currency.code} -{" "}
                <input value={name} onChange={(e) => setName(e.target.value)} />
                <input 
                    type="number"
                    value={decimalPlaces}
                    onChange={(e) => setDecimalPlaces(e.target.value)}
                />
                <button onClick={save}>Save</button>
                <button onClick={() => setIsEditing(false)}>Cancel</button>
            </li>
        );
    }

    return (
        <li>
            {currency.code} - {currency.name} ({currency.kind}, {currency.decimal_places} decimals)
            {!currency.is_active && " [inactive]"}
            {currency.is_active && (
                <>
                    <button onClick={startEdit}>Edit</button>
                    <button onClick={() => onDeactivate(currency.code)}>Deactivate</button>
                </>
            )}
        </li>
    );
}

export default CurrencyRow;
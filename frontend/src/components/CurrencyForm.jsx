import { useState } from "react";
import { createCurrency } from "../api/client";

function CurrencyForm({ onCreated }) {
    const [code, setCode] = useState("");
    const [name, setName] = useState("");
    const [kind, setKind] = useState("fiat");
    const [decimalPlaces, setDecimalPlaces] = useState(2);
    const [error, setError] = useState(null);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await createCurrency({
                code: code,
                name: name,
                kind: kind,
                decimal_places: Number(decimalPlaces),
            });
            setCode("");
            setName("");
            setKind("fiat");
            setDecimalPlaces(2);
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form onSubmit={handleSubmit}>
            <h3>Add currency</h3>
            {error && <div style={{ color: "red" }}>{error}</div>}
            <input
                placeholder="Code (USD)"
                value={code}
                onChange={(e) => setCode(e.target.value)}
            />
            <input
                placeholder="Name (US Dollar)"
                value={name}
                onChange={(e) => setName(e.target.value)}
            />
            <select value={kind} onChange={(e) => setKind(e.target.value)}>
                <option value="fiat">fiat</option>
                <option value="crypto">crypto</option>
            </select>
            <input
                type="number"
                placeholder="Decimals"
                value={decimalPlaces}
                onChange={(e) => setDecimalPlaces(e.target.value)}
            />
            <button type="submit">Create</button>
        </form>
    );
}

export default CurrencyForm;
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
        <form className="card" onSubmit={handleSubmit}>
            <h5 className="form-title">Add currency</h5>
            {error && <div className="error">{error}</div>}
            <div className="form-grid">
                <div className="field">
                    <label>Code</label>
                    <input
                        className="input"
                        placeholder="USD"
                        value={code}
                        onChange={(e) => setCode(e.target.value)}
                    />
                </div>
                <div className="field">
                    <label>Name</label>
                    <input
                        className="input"
                        placeholder="US Dollar"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                    />
                </div>
                <div className="field">
                    <label>Kind</label>
                    <select className="input" value={kind} onChange={(e) => setKind(e.target.value)}>
                        <option value="fiat">fiat</option>
                        <option value="crypto">crypto</option>
                    </select>
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
                <button type="submit" className="btn btn-primary">Create</button>
            </div>
        </form>
    );
}

export default CurrencyForm;
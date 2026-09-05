import { useState } from "react"

function RequisiteRow({ requisite, label, onClose }) {
    const [validTo, setValidTo] = useState("");
    const [error, setError] = useState(null);

    async function handleClose() {
        if (!validTo) {
            setError("Pick a date");
            return;
        }
        setError(null);
        try {
            await onClose(requisite.id, `${validTo}T00:00:00Z`);
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <li className={requisite.valid_to ? "row row-inactive" : "row"}>
            <span>{label}</span>
            <span className="row-spacer" />
            {requisite.valid_to ? (
                <span className="badge badge-closed">
                    closed {new Date(requisite.valid_to).toLocaleDateString()}
                </span>
            ) : (
                <div className="row-actions">
                    <input
                        className="input"
                        type="date"
                        style={{ width: "auto" }}
                        value={validTo}
                        onChange={(e) => setValidTo(e.target.value)}
                    />
                    <button className="btn btn-ghost btn-sm" onClick={handleClose}>Close</button>
                    {error && <span className="error" style={{ margin: 0, padding: "4px 8px" }}>{error}</span>}
                </div>
            )}
        </li>
    );
}

export default RequisiteRow;
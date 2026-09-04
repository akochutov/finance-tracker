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
        <li>
            {label}
            {requisite.valid_to ? (
                ` [closed ${new Date(requisite.valid_to).toLocaleDateString()}]`
            ) : (
                <>
                    {" "}
                    <input
                        type="date"
                        value={validTo}
                        onChange={(e) => setValidTo(e.target.value)}
                    />
                    <button onClick={handleClose}>Close</button>
                    {error && <span style={{ color: "red" }}> {error}</span>}
                </>
            )}
        </li>
    );
}

export default RequisiteRow;
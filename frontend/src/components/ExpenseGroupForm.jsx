import { useState } from "react";
import { createExpenseGroup } from "../api/client";

function ExpenseGroupForm({ onCreated }) {
    const [name, setName] = useState("");
    const [error, setError] = useState(null);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await createExpenseGroup(name);
            setName("");
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form className="card" onSubmit={handleSubmit}>
            <h5 className="form-title">Add group</h5>
            {error && <div className="error">{error}</div>}
            <div className="form-grid">
                <div className="field">
                    <label>Name</label>
                    <input className="input" placeholder="Продукты" value={name} onChange={(e) => setName(e.target.value)} />
                </div>
            </div>
            <div className="form-actions">
                <button type="submit" className="btn btn-primary">Create</button>
            </div>
        </form>
    );
}

export default ExpenseGroupForm;
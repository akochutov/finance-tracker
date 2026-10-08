import { useState } from "react";
import { createExpenseCategory } from "../api/client";

function ExpenseCategoryForm({ groups, onCreated }) {
    const [groupId, setGroupId] = useState("");
    const [name, setName] = useState("");
    const [error, setError] = useState(null);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        if (!groupId) {
            setError("Choose a group");
            return;
        }
        try {
            await createExpenseCategory(groupId, name);
            setName("");
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form className="card" onSubmit={handleSubmit}>
            <h5 className="form-title">Add category</h5>
            {error && <div className="error">{error}</div>}
            <div className="form-row">
                <div className="field">
                    <label>Group</label>
                    <select className="input" value={groupId} onChange={(e) => setGroupId(e.target.value)}>
                        <option value="">- Group -</option>
                        {groups.map((g) => (
                            <option key={g.id} value={g.id}>{g.name}</option>
                        ))}
                    </select>
                </div>
                <div className="field">
                    <label>Name</label>
                    <input className="input" placeholder="Молоко" value={name} onChange={(e) => setName(e.target.value)} />
                </div>
            </div>
            <div className="form-actions">
                <button type="submit" className="btn btn-primary">Create</button>
            </div>
        </form>
    );
}

export default ExpenseCategoryForm;
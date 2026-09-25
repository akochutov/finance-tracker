import { useState } from "react";
import ExpenseCategoryRow from "./ExpenseCategoryRow";

function ExpenseGroupRow({ group, groups, groupActions, categoryActions }) {
    const [expanded, setExpanded] = useState(false);
    const [isEditing, setIsEditing] = useState(false);
    const [name, setName] = useState(group.name);

    function startEdit() {
        setName(group.name);
        setIsEditing(true);
    }

    async function save() {
        const ok = await groupActions.save(group.id, name);
        if (ok) {
            setIsEditing(false);
        }
    }

    const activeCount = group.categories.filter((c) => c.is_active).length;

    return (
        <li className="expense-group-item">
            {isEditing ? (
                <div className="row-editing">
                    <div className="form-grid">
                        <div className="field">
                            <label>Group name</label>
                            <input className="input" value={name} onChange={(e) => setName(e.target.value)} />
                        </div>
                        <div className="row-actions" style={{ justifyContent: "flex-start" }}>
                            <button className="btn btn-primary btn-sm" onClick={save}>Save</button>
                            <button className="btn btn-secondary btn-sm" onClick={() => setIsEditing(false)}>Cancel</button>
                        </div>
                    </div>
                </div>
            ) : (
                <div className={group.is_active ? "row-expense-group" : "row-expense-group row-inactive"}>
                    <button type="button" className="group-toggle" onClick={() => setExpanded(!expanded)}>
                        <span className="group-toggle-icon">{expanded ? "▼" : "▶"}</span>
                        <span className="row-key">{group.name}</span>
                    </button>
                    <span className="row-meta">
                        {activeCount} of {group.categories.length} active
                    </span>
                    <div className="row-actions">
                        {group.is_active ? (
                            <>
                                <button className="btn btn-secondary btn-sm" onClick={startEdit}>Edit</button>
                                <button className="btn btn-ghost btn-sm" onClick={() => groupActions.deactivate(group.id)}>Deactivate</button>
                            </>
                        ) : (
                            <>
                                <span className="badge badge-inactive">inactive</span>
                                <button className="btn btn-secondary btn-sm" onClick={() => groupActions.activate(group.id)}>Activate</button>
                            </>
                        )}
                    </div>
                </div>
            )}

            {expanded && (
                <ul className="list-nested">
                    {group.categories.length === 0 && (
                        <li className="row-empty">No categories yet</li>
                    )}
                    {group.categories.map((c) => (
                        <ExpenseCategoryRow 
                            key={c.id}
                            category={c}
                            groups={groups}
                            actions={categoryActions}
                        />
                    ))}
                </ul>
            )}
        </li>
    );
}

export default ExpenseGroupRow;
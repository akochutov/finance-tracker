import { useState } from "react";

function ExpenseCategoryRow({ category, groups, actions }) {
    const [isEditing, setIsEditing] = useState(false);
    const [name, setName] = useState(category.name);
    const [groupId, setGroupId] = useState(category.group_id);

    const groupOptions = groups.filter((g) => g.is_active || g.id === category.group_id);

    function startEdit() {
        setName(category.name);
        setGroupId(category.group_id);
        setIsEditing(true);
    }

    async function save() {
        const ok = await actions.save(category.id, groupId, name);
        if (ok) {
            setIsEditing(false);
        }
    }

    if (isEditing) {
        return (
            <li className="row-editing">
                <div className="form-grid">
                    <div className="field">
                        <label>Group</label>
                        <select className="input" value={groupId} onChange={(e) => setGroupId(e.target.value)}>
                            {groupOptions.map((g) => (
                                <option key={g.id} value={g.id}>{g.name}</option>
                            ))}
                        </select>
                    </div>
                    <div className="field">
                        <label>Name</label>
                        <input className="input" value={name} onChange={(e) => setName(e.target.value)} />
                    </div>
                    <div className="row-actions" style={{ justifyContent: "flex-start" }}>
                        <button className="btn btn-primary btn-sm" onClick={save}>Save</button>
                        <button className="btn btn-secondary btn-sm" onClick={() => setIsEditing(false)}>Cancel</button>
                    </div>
                </div>
            </li>
        );
    }

    return (
        <li className={category.is_active ? "row-expense-category" : "row-expense-category row-inactive"}>
            <span>{category.name}</span>
            <div className="row-actions">
                <label className="checkbox" title="Count this category on the expenses dashboard">
                    <input
                        type="checkbox"
                        checked={category.include_in_dashboard}
                        onChange={(e) => actions.setDashboard(category.id, e.target.checked)}
                    />
                    Dashboard
                </label>
                {category.is_active ? (
                    <>
                        <button className="btn btn-secondary btn-sm" onClick={startEdit}>Edit</button>
                        <button className="btn btn-ghost btn-sm" onClick={() => actions.deactivate(category.id)}>Deactivate</button>
                    </>
                ) : (
                    <>
                        <span className="badge badge-inactive">inactive</span>
                        <button className="btn btn-secondary btn-sm" onClick={() => actions.activate(category.id)}>Activate</button>
                    </>
                )}
            </div>
        </li>
    );
}

export default ExpenseCategoryRow;
import { useState, useEffect } from "react";
import { 
    getExpenseGroups, updateExpenseGroup, deactivateExpenseGroup, activateExpenseGroup,
    updateExpenseCategory, deactivateExpenseCategory, activateExpenseCategory,
} from "../api/client";
import ExpenseGroupForm from "./ExpenseGroupForm";
import ExpenseCategoryForm from "./ExpenseCategoryForm";
import ExpenseGroupRow from "./ExpenseGroupRow";

function ExpenseCategoriesPage() {
    const [groups, setGroups] = useState([]);
    const [error, setError] = useState(null);

    async function loadGroups() {
        try {
            const data = await getExpenseGroups();
            setGroups(data);
        } catch (err) {
            setError(err.message);
        }
    }

    async function run(action) {
        setError(null);
        try {
            await action();
            await loadGroups();
            return true;
        } catch (err) {
            setError(err.message);
            window.scrollTo({ top: 0, behavior: "smooth" });
            return false;
        }
    }

    const groupActions = {
        save: (id, name) => run(() => updateExpenseGroup(id, name)),
        deactivate: (id) => run(() => deactivateExpenseGroup(id)),
        activate: (id) => run(() => activateExpenseGroup(id)),
    };

    const categoryActions = {
        save: (id, groupId, name) => run(() => updateExpenseCategory(id, groupId, name)),
        deactivate: (id) => run(() => deactivateExpenseCategory(id)),
        activate: (id) => run(() => activateExpenseCategory(id)),
    };

    useEffect(() => {
        loadGroups();
    }, []);

    const activeGroups = groups.filter((g) => g.is_active);
    const categoryCount = groups.reduce((sum, g) => sum + g.categories.length, 0);

    return (
        <div>
            <div className="page-header">
                <h1>Expense categories</h1>
                <p>Groups and their categories used to classify expenses.</p>
            </div>
            {error && <div className="error">{error}</div>}
            <ExpenseGroupForm onCreated={loadGroups} />
            <ExpenseCategoryForm groups={activeGroups} onCreated={loadGroups} />
            <div className="list-header">
                <h5>All groups</h5>
                <span className="row-meta">{groups.length} groups, {categoryCount} categories</span>
            </div>
            <ul className="list">
                {groups.map((g) => (
                    <ExpenseGroupRow
                        key={g.id}
                        group={g}
                        groups={groups}
                        groupActions={groupActions}
                        categoryActions={categoryActions}
                    />
                ))}
            </ul>
        </div>
    );
}

export default ExpenseCategoriesPage;
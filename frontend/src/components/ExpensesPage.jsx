import { useState, useEffect } from "react";
import { getExpenses, deleteExpense, getExpenseGroups, getCurrencies } from "../api/client";
import ExpenseForm from "./ExpenseForm";
import ExpenseRow from "./ExpenseRow";
import { currentMonth, shiftMonth, monthLabel, monthRange, sumByCurrency, formatAmount } from "./expenseUtils";

function ExpensesPage() {
    const [expenses, setExpenses] = useState([]);
    const [groups, setGroups] = useState([]);
    const [currencies, setCurrencies] = useState([]);
    const [loaded, setLoaded] = useState(false);
    const [month, setMonth] = useState(currentMonth());
    const [editing, setEditing] = useState(null);
    const [reloadKey, setReloadKey] = useState(0);
    const [error, setError] = useState(null);

    useEffect(() => {
        async function loadOptions() {
            try {
                setGroups(await getExpenseGroups());
                setCurrencies(await getCurrencies());
                setLoaded(true);
            } catch (err) {
                setError(err.message);
            }
        }
        loadOptions();
    }, []);

    useEffect(() => {
        let ignore = false;
        const [from, to] = monthRange(month);
        getExpenses(from, to)
            .then((data) => {
                if (!ignore) setExpenses(data);
            })
            .catch((err) => {
                if (!ignore) setError(err.message);
            });
        return () => {
            ignore = true;
        };
    }, [month, reloadKey]);

    function reloadExpenses() {
        setReloadKey((k) => k + 1);
    }

    const categoriesById = {};
    for (const g of groups) {
        for (const c of g.categories) {
            categoriesById[c.id] = `${g.name} → ${c.name}`;
        }
    }

    const defaultCurrency = expenses.length > 0
        ? expenses[0].currency
        : (currencies.some((c) => c.code === "AMD") ? "AMD" : currencies[0]?.code || "");

    function handleSaved() {
        setEditing(null);
        reloadExpenses();
    }

    function handleEdit(expense) {
        setEditing(expense);
        window.scrollTo({ top: 0, behavior: "smooth" });
    }

    async function handleDelete(expense) {
        if (!window.confirm("Delete this expense? This cannot be undone.")) {
            return;
        }
        setError(null);
        try {
            await deleteExpense(expense.id);
            if (editing && editing.id === expense.id) {
                setEditing(null);
            }
            reloadExpenses();
        } catch (err) {
            setError(err.message);
        }
    }

    const visible = expenses;
    const totals = Object.entries(sumByCurrency(visible))
        .map(([code, sum]) => `${formatAmount(sum)} ${code}`)
        .join(" · ");

    return (
        <div>
            <div className="page-header">
                <h1>Expenses</h1>
                <p>Receipts and regular payments, line by line.</p>
            </div>
            {error && <div className="error">{error}</div>}

            {loaded && (
                <ExpenseForm
                    key={editing ? editing.id : "new"}
                    expense={editing}
                    groups={groups}
                    currencies={currencies}
                    defaultCurrency={defaultCurrency}
                    onSaved={handleSaved}
                    onCancel={() => setEditing(null)}
                />
            )}

            <div className="list-header">
                <div className="month-nav">
                    <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, -1))}>←</button>
                    <h5>{monthLabel(month)}</h5>
                    <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, 1))}>→</button>
                </div>
                <span className="row-meta">
                    {visible.length} expenses{totals && ` · ${totals}`}
                </span>
            </div>
            <ul className="list">
                {visible.map((e) => (
                    <ExpenseRow
                        key={e.id}
                        expense={e}
                        categoriesById={categoriesById}
                        onEdit={handleEdit}
                        onDelete={handleDelete}
                    />
                ))}
            </ul>
            {loaded && visible.length === 0 && (
                <div className="list-empty">No expenses in this month.</div>
            )}
        </div>
    );
}

export default ExpensesPage;
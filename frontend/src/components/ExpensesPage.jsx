import { useState, useEffect } from "react";
import { getExpenses, deleteExpense, getExpenseGroups, getCurrencies, getExpenseSuggestions } from "../api/client";
import ExpenseForm from "./ExpenseForm";
import ExpenseRow from "./ExpenseRow";
import { currentMonth, shiftMonth, monthRange, sumByCurrency, formatAmount } from "./expenseUtils";
import { periodRange, stepMonths } from "./dashboardUtils";

const MODES = [
    ["month", "Month"],
    ["quarter", "Quarter"],
    ["year", "Year"],
    ["custom", "Custom"],
];

function ExpensesPage() {
    const [expenses, setExpenses] = useState([]);
    const [groups, setGroups] = useState([]);
    const [currencies, setCurrencies] = useState([]);
    const [loaded, setLoaded] = useState(false);
    const [mode, setMode] = useState("month");
    const [month, setMonth] = useState(currentMonth());
    const [custom, setCustom] = useState(() => {
        const [from, to] = monthRange(currentMonth());
        return { from, to };
    });
    const [query, setQuery] = useState("");
    const [editing, setEditing] = useState(null);
    const [reloadKey, setReloadKey] = useState(0);
    const [suggestions, setSuggestions] = useState([]);
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

    // Same period logic as the expenses dashboard.
    const range = mode === "custom" ? custom : periodRange(mode, month);
    const rangeValid = Boolean(range.from && range.to && range.to >= range.from);

    useEffect(() => {
        if (!rangeValid) return;
        let ignore = false;
        getExpenses(range.from, range.to)
            .then((data) => {
                if (!ignore) setExpenses(data);
            })
            .catch((err) => {
                if (!ignore) setError(err.message);
            });
        return () => {
            ignore = true;
        };
    }, [range.from, range.to, rangeValid, reloadKey]);

    useEffect(() => {
        let ignore = false;
        getExpenseSuggestions()
            .then((data) => {
                if (ignore) return;
                setSuggestions(data.map((s) => ({ ...s, lower: s.description.toLowerCase() })));
            })
            .catch((err) => {
                console.warn("load suggestions:", err);
            });
        return () => {
            ignore = true;
        };
    }, [reloadKey]);

    // Custom starts from the period on screen, so "Q3 → Custom" begins with Q3.
    function changeMode(next) {
        if (next === "custom" && mode !== "custom") {
            setCustom({ from: range.from, to: range.to });
        }
        setMode(next);
    }

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

    // Text filter over the loaded month: receipt note or any line description.
    // An invalid custom range shows nothing rather than the previous period.
    const loadedList = rangeValid ? expenses : [];
    const q = query.trim().toLowerCase();
    const visible = q === ""
        ? loadedList
        : loadedList.filter((e) =>
            (e.note || "").toLowerCase().includes(q) ||
            e.items.some((it) => (it.description || "").toLowerCase().includes(q)));
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
                    suggestions={suggestions}
                    categoriesById={categoriesById}
                    onSaved={handleSaved}
                    onCancel={() => setEditing(null)}
                />
            )}

            <div className="list-header expense-toolbar">
                <div className="xd-controls">
                    <div className="xd-segmented">
                        {MODES.map(([key, label]) => (
                            <button
                                type="button"
                                key={key}
                                className={mode === key ? "active" : ""}
                                onClick={() => changeMode(key)}
                            >
                                {label}
                            </button>
                        ))}
                    </div>

                    {mode === "custom" ? (
                        <div className="xd-custom">
                            <input className="input" type="date" value={custom.from} onChange={(e) => setCustom({ ...custom, from: e.target.value })} />
                            <span className="row-meta">–</span>
                            <input className="input" type="date" value={custom.to} onChange={(e) => setCustom({ ...custom, to: e.target.value })} />
                        </div>
                    ) : (
                        <div className="month-nav">
                            <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, -stepMonths[mode]))}>←</button>
                            <h5>{range.label}</h5>
                            <button className="btn btn-secondary btn-sm" onClick={() => setMonth(shiftMonth(month, stepMonths[mode]))}>→</button>
                        </div>
                    )}
                </div>
                <input
                    type="search"
                    className="input list-search"
                    placeholder="Filter by text…"
                    aria-label="Filter expenses by text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    onKeyDown={(e) => {
                        if (e.key === "Escape") setQuery("");
                    }}
                />
                <span className="row-meta">
                    {q ? `${visible.length} of ${loadedList.length}` : visible.length} expenses{totals && ` · ${totals}`}
                </span>
            </div>
            {!rangeValid && <div className="error">The end of the period is before its start.</div>}
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
            {loaded && rangeValid && visible.length === 0 && (
                <div className="list-empty">
                    {q && loadedList.length > 0
                        ? `Nothing matches "${query.trim()}" in this period.`
                        : "No expenses in this period."}
                </div>
            )}
        </div>
    );
}

export default ExpensesPage;
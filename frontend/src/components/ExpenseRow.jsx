import { useState } from "react";
import { formatISODate, formatAmount, expenseTotal } from "./Expenseutils";

function describeQuantity(it) {
    const qty = Number(it.quantity);
    const discount = Number(it.discount);
    let text = `${qty} × ${formatAmount(Number(it.price))}`;
    if (discount) {
        text += ` − ${formatAmount(discount)}`;
    }
    return text;
}

function describePeriod(it) {
    if (!it.period_from) {
        return "";
    }
    return `${formatISODate(it.period_from)} – ${formatISODate(it.period_to)}`;
}

function ExpenseRow({ expense, categoriesById, onEdit, onDelete }) {
    const [expanded, setExpanded] = useState(false);

    const summary = expense.note || expense.items.map((it) => it.description).join(", ");
    const count = expense.items.length;

    return (
        <li className="expense-item">
            <div className="row-expense" onClick={() => setExpanded(!expanded)}>
                <span className="row-meta">{formatISODate(expense.occurred_on)}</span>
                <span className="row-key mono">{formatAmount(expenseTotal(expense))} {expense.currency}</span>
                <span className="expense-summary" title={summary}>{summary}</span>
                <span className="row-meta">{count} {count === 1 ? "line" : "lines"}</span>
                <span>{expense.payment_type && <span className="badge">{expense.payment_type}</span>}</span>
                {/* stopPropagation: button clicks must not toggle the row */}
                <div className="row-actions" onClick={(e) => e.stopPropagation()}>
                    <button className="btn btn-secondary btn-sm" onClick={() => onEdit(expense)}>Edit</button>
                    <button className="btn btn-ghost btn-sm" onClick={() => onDelete(expense)}>Delete</button>
                </div>
            </div>

            {expanded && (
                <div className="expense-details">
                    {expense.items.map((it) => (
                        <div className="expense-detail-line" key={it.id}>
                            <span>{it.description}</span>
                            <span className="row-meta">{categoriesById[it.category_id] || "unknown category"}</span>
                            <span className="row-meta mono">{describeQuantity(it)}</span>
                            <span className="row-meta">{describePeriod(it)}</span>
                            <span className="mono num">{formatAmount(Number(it.amount))}</span>
                        </div>
                    ))}
                </div>
            )}
        </li>
    );
}

export default ExpenseRow;
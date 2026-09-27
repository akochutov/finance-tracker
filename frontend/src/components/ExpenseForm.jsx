import { useState } from "react";
import { createExpense, updateExpense } from "../api/client";
import CategorySelect from "./CategorySelect";
import { todayISO, lineTotal, formatAmount } from "./Expenseutils";

function emptyLine() {
    return {
        key: crypto.randomUUID(),
        description: "",
        categoryId: "",
        price: "",
        quantity: "1",
        discount: "",
        periodFrom: "",
        periodTo: "",
    };
}

function lineFromItem(it) {
    return {
        key: it.id,
        description: it.description,
        categoryId: it.category_id,
        price: String(it.price),
        quantity: String(Number(it.quantity)),
        discount: Number(it.discount) ? String(it.discount) : "",
        periodFrom: it.period_from ? it.period_from.slice(0, 10) : "",
        periodTo: it.period_to ? it.period_to.slice(0, 10) : "",
    };
}

function ExpenseForm({ expense, groups, currencies, defaultCurrency, onSaved, onCancel }) {
    const isEdit = expense !== null;

    const [occurredOn, setOccurredOn] = useState(isEdit ? expense.occurred_on.slice(0, 10) : todayISO());
    const [currency, setCurrency] = useState(isEdit ? expense.currency : defaultCurrency);
    const [paymentType, setPaymentType] = useState(isEdit ? expense.payment_type || "" : "");
    const [note, setNote] = useState(isEdit ? expense.note || "" : "");
    const [lines, setLines] = useState(isEdit ? expense.items.map(lineFromItem) : [emptyLine()]);
    const [withPeriod, setWithPeriod] = useState(isEdit && expense.items.some((it) => it.period_from));
    const [error, setError] = useState(null);
    const [saving, setSaving] = useState(false);

    const keepIds = new Set(isEdit ? expense.items.map((it) => it.category_id) : []);
    const currencyOptions = currencies.filter((c) => c.is_active || c.code === currency);

    function updateLine(key, field, value) {
        setLines((prev) => prev.map((l) => (l.key === key ? { ...l, [field]: value } : l)));
    }

    function addLine() {
        setLines((prev) => [...prev, emptyLine()]);
    }

    function removeLine(key) {
        setLines((prev) => (prev.length > 1 ? prev.filter((l) => l.key !== key) : prev));
    }

    const total = lines.reduce((sum, l) => sum + lineTotal(l.price, l.quantity, l.discount), 0);

    function toPayload() {
        return {
            occurred_on: occurredOn,
            currency: currency,
            payment_type: paymentType || null,
            note: note || null,
            items: lines.map((l) => ({
                description: l.description,
                category_id: l.categoryId || null,
                price: l.price === "" ? null : l.price,
                quantity: l.quantity === "" ? null : l.quantity,
                discount: l.discount === "" ? null : l.discount,
                period_from: withPeriod ? l.periodFrom : "",
                period_to: withPeriod ? l.periodTo : "",
            })),
        };
    }

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        setSaving(true);
        try {
            if (isEdit) {
                await updateExpense(expense.id, toPayload());
            } else {
                await createExpense(toPayload());
                setLines([emptyLine()]);
                setNote("");
            }
            onSaved();
        } catch (err) {
            setError(err.message);
        } finally {
            setSaving(false);
        }
    }

    function blockEnter(e) {
        if (e.key === "Enter") {
            e.preventDefault();
        }
    }

    return (
        <form className="card" onSubmit={handleSubmit}>
            <h5 className="form-title">{isEdit ? "Edit expense" : "Add expense"}</h5>
            {error && <div className="error">{error}</div>}

            <div className="form-grid">
                <div className="field">
                    <label>Date</label>
                    <input className="input" type="date" value={occurredOn} onChange={(e) => setOccurredOn(e.target.value)} />
                </div>
                <div className="field">
                    <label>Currency</label>
                    <select className="input" value={currency} onChange={(e) => setCurrency(e.target.value)}>
                        <option value="">- Currency -</option>
                        {currencyOptions.map((c) => (
                            <option key={c.code} value={c.code}>{c.code}</option>
                        ))}
                    </select>
                </div>
                <div className="field">
                    <label>Payment type</label>
                    <select className="input" value={paymentType} onChange={(e) => setPaymentType(e.target.value)}>
                        <option value="">-</option>
                        <option value="bank">bank</option>
                        <option value="crypto">crypto</option>
                        <option value="cash">cash</option>
                    </select>
                </div>
                <div className="field">
                    <label>Note</label>
                    <input className="input" placeholder="Optional" value={note} onChange={(e) => setNote(e.target.value)} />
                </div>
            </div>

            <div className={withPeriod ? "expense-lines with-period" : "expense-lines"} onKeyDown={blockEnter}>
                <div className="expense-line expense-line-head">
                    <span className="line-no">#</span>
                    <span>Description</span>
                    <span>Category</span>
                    <span>Price</span>
                    <span>Qty</span>
                    <span>Discount</span>
                    {withPeriod && (
                        <>
                            <span>Period from</span>
                            <span>Period to</span>
                        </>
                    )}
                    <span className="num">Total</span>
                    <span></span>
                </div>

                {lines.map((l, i) => (
                    <div className="expense-line" key={l.key}>
                        <span className="line-no">{i + 1}</span>
                        <input className="input" value={l.description} onChange={(e) => updateLine(l.key, "description", e.target.value)} />
                        <CategorySelect
                            groups={groups}
                            value={l.categoryId}
                            onChange={(v) => updateLine(l.key, "categoryId", v)}
                            keepIds={keepIds}
                        />
                        <input className="input" type="number" step="any" min="0" value={l.price} onChange={(e) => updateLine(l.key, "price", e.target.value)} />
                        <input className="input" type="number" step="any" min="0" value={l.quantity} onChange={(e) => updateLine(l.key, "quantity", e.target.value)} />
                        <input className="input" type="number" step="any" min="0" placeholder="0" value={l.discount} onChange={(e) => updateLine(l.key, "discount", e.target.value)} />
                        {withPeriod && (
                            <>
                                <input className="input" type="date" value={l.periodFrom} onChange={(e) => updateLine(l.key, "periodFrom", e.target.value)} />
                                <input className="input" type="date" value={l.periodTo} onChange={(e) => updateLine(l.key, "periodTo", e.target.value)} />
                            </>
                        )}
                        <span className="line-total mono num">{formatAmount(lineTotal(l.price, l.quantity, l.discount))}</span>
                        <button
                            type="button"
                            className="btn btn-ghost btn-sm"
                            title="Remove line"
                            disabled={lines.length === 1}
                            onClick={() => removeLine(l.key)}
                        >
                            ×
                        </button>
                    </div>
                ))}
            </div>

            <div className="expense-form-footer">
                <button type="button" className="btn btn-secondary btn-sm" onClick={addLine}>+ Line</button>
                <label className="checkbox">
                    <input type="checkbox" checked={withPeriod} onChange={(e) => setWithPeriod(e.target.checked)} />
                    With period
                </label>
                <span className="expense-total">
                    Total: <b className="mono">{formatAmount(total)} {currency}</b>
                </span>
            </div>

            <div className="form-actions">
                {isEdit && (
                    <button type="button" className="btn btn-secondary" onClick={onCancel}>Cancel</button>
                )}
                <button type="submit" className="btn btn-primary" disabled={saving}>
                    {isEdit ? "Save" : "Create"}
                </button>
            </div>
        </form>
    );
}

export default ExpenseForm;
function pad(n) {
    return String(n).padStart(2, "0");
}

export function todayISO() {
    const d = new Date();
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function currentMonth() {
    return todayISO().slice(0, 7);
}

export function shiftMonth(month, delta) {
    const [y, m] = month.split("-").map(Number);
    const d = new Date(y, m - 1 + delta, 1);
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}`;
}

export function monthLabel(month) {
    const [y, m] = month.split("-").map(Number);
    return new Date(y, m - 1, 1).toLocaleDateString(undefined, { month: "long", year: "numeric" });
}

export function formatISODate(iso) {
    const [y, m, d] = iso.slice(0, 10).split("-");
    return `${d}.${m}.${y}`;
}

export function lineTotal(price, quantity, discount) {
    const p = Number(price) || 0;
    const q = quantity === "" ? 1 : Number(quantity) || 0;
    const d = Number(discount) || 0;
    return p * q - d;
}

export function formatAmount(n) {
    return n.toLocaleString(undefined, { maximumFractionDigits: 2 });
}

export function expenseTotal(expense) {
    return expense.items.reduce((sum, it) => sum + Number(it.amount), 0);
}

export function sumByCurrency(expenses) {
    const totals = {};
    for (const e of expenses) {
        totals[e.currency] = (totals[e.currency] || 0) + expenseTotal(e);
    }
    return totals;
}
import { monthRange, shiftMonth, monthLabel, formatISODate } from "./expenseUtils";

function pad(n) {
    return String(n).padStart(2, "0");
}

export const stepMonths = { month: 1, quarter: 3, year: 12 };

export function periodRange(mode, month) {
    const [y, m] = month.split("-").map(Number);
    if (mode === "quarter") {
        const startMonth = Math.floor((m - 1) / 3) * 3 + 1;
        const start = `${y}-${pad(startMonth)}`;
        const [, to] = monthRange(shiftMonth(start, 2));
        return { from: `${start}-01`, to, label: `Q${(startMonth + 2) / 3} ${y}` };
    }
    if (mode === "year") {
        return { from: `${y}-01-01`, to: `${y}-12-31`, label: String(y) };
    }
    const [from, to] = monthRange(month);
    return { from, to, label: monthLabel(month) };
}

export function previousLabel(mode, previous) {
    if (mode === "custom") {
        return `${formatISODate(previous.from)} – ${formatISODate(previous.to)}`;
    }
    return periodRange(mode, previous.from.slice(0, 7)).label;
}

export function formatWhole(n) {
    return Math.round(Number(n)).toLocaleString();
}

export function formatCompact(n) {
    const v = Number(n);
    const a = Math.abs(v);
    const trim = (x) => x.toFixed(Math.abs(x) >= 100 ? 0 : Math.abs(x) >= 10 ? 1 : 2).replace(/\.?0+$/, "");
    if (a >= 1e6) return trim(v / 1e6) + "M";
    if (a >= 1e3) return trim(v / 1e3) + "k";
    return String(Math.round(v));
}

export function percentChange(current, previous) {
    const cur = Number(current);
    const prev = Number(previous);
    if (!(prev > 0)) return null;
    return ((cur - prev) / prev) * 100;
}

export function niceMax(value) {
    if (!(value > 0)) return 1;
    const power = 10 ** Math.floor(Math.log10(value));
    const n = value / power;
    const nice = n <= 1 ? 1 : n <= 2 ? 2 : n <= 2.5 ? 2.5 : n <= 5 ? 5 : 10;
    return nice * power;
}

export function monthShort(key, first) {
    const [y, m] = key.split("-").map(Number);
    const name = new Date(y, m - 1, 1).toLocaleDateString(undefined, { month: "short" });
    return m === 1 || first ? `${name} ${String(y).slice(2)}` : name;
}

export const GROUP_COLORS = ["#2f49b0", "#e8590c", "#0b7285", "#f59f00", "#7950f2"];
export const OTHER_COLOR = "#ced4da";
export const PLAIN_BAR_COLOR = "#868e96";
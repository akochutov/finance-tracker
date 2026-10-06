export const ZONES_SINGLE = "single";
export const ZONES_DAY_NIGHT = "day_night";

export function isDayNight(account) {
    return account && account.zones === ZONES_DAY_NIGHT;
}

export function billingMonthKey(date) {
    const [y, m] = date.slice(0, 10).split("-").map(Number);
    const d = new Date(Date.UTC(y, m - 2, 1));
    return `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, "0")}`;
}

export function monthLabel(key) {
    const [y, m] = key.split("-").map(Number);
    return new Date(y, m - 1, 1).toLocaleDateString(undefined, { month: "short", year: "numeric" });
}

export function formatReading(value) {
    return Number(value).toLocaleString(undefined, { maximumFractionDigits: 3 });
}

export function sameAmount(a, b) {
    return Math.round(a * 1000) === Math.round(b * 1000);
}
export const REGISTERS_SINGLE = "single";
export const REGISTERS_DAY_NIGHT = "day_night";

export const ZONE_SINGLE = "single";
export const ZONE_DAY = "day";
export const ZONE_NIGHT = "night";

export function isDayNight(meter) {
    return Boolean(meter) && meter.registers === REGISTERS_DAY_NIGHT;
}

export function zonesOf(meter) {
    return isDayNight(meter) ? [ZONE_DAY, ZONE_NIGHT] : [ZONE_SINGLE];
}

export function zoneLabel(zone) {
    if (zone === ZONE_DAY) return "Day";
    if (zone === ZONE_NIGHT) return "Night";
    return "";
}

export function registerKey(meterId, zone) {
    return `${meterId}:${zone}`;
}

export function groupByDate(readings) {
    const groups = new Map();
    for (const r of readings) {
        const day = r.taken_on.slice(0, 10);
        if (!groups.has(day)) groups.set(day, { taken_on: r.taken_on, is_initial: false, byZone: {} });
        const g = groups.get(day);
        g.byZone[r.zone] = r;
        if (r.is_initial) g.is_initial = true;
    }
    return [...groups.values()].sort((a, b) => b.taken_on.localeCompare(a.taken_on));
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
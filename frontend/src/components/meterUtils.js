const ZONE_ORDER = { day: 0, night: 1, single: 2 };

export function meterZones(meter) {
    return meter.registers
        .map((r) => r.zone)
        .sort((a, b) => ZONE_ORDER[a] - ZONE_ORDER[b]);
}

export function zoneLabel(zone) {
    return { day: "Day", night: "Night", single: "Value" }[zone] || zone;
}

export function isDual(meter) {
    return meter.registers.some((r) => r.zone === "day");
}

export function groupReadings(meter, readings) {
    const zoneOf = {};
    for (const reg of meter.registers) zoneOf[reg.id] = reg.zone;

    const byDate = new Map();
    for (const r of readings) {
        const date = r.taken_on.slice(0, 10);
        if (!byDate.has(date)) byDate.set(date, { date, byZone: {} });
        byDate.get(date).byZone[zoneOf[r.register_id]] = r;
    }
    return [...byDate.values()].sort((a, b) => b.date.localeCompare(a.date));
}

export function consumption(rows, i, zone) {
    const current = rows[i].byZone[zone];
    if (!current) return null;
    for (let j = i + 1; j < rows.length; j++) {
        const older = rows[j].byZone[zone];
        if (older) return Number(current.value) - Number(older.value);
    }
    return null;
}

export function billingMonth(date) {
    const [y, m] = date.split("-").map(Number);
    return new Date(y, m - 2, 1).toLocaleDateString(undefined, { month: "short", year: "numeric" });
}

export function formatReading(value) {
    return Number(value).toLocaleString(undefined, { maximumFractionDigits: 3 });
}
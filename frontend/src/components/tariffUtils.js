export const MODE_WHOLE = "whole";
export const MODE_PROGRESSIVE = "progressive";

export const MODE_LABELS = {
    [MODE_WHOLE]: "Whole month at its tier",
    [MODE_PROGRESSIVE]: "Progressive",
};

export const ZONES = ["single", "day", "night"];

export function emptyTiers() {
    return [{ upTo: "", price: "" }];
}

export function tiersFromTariff(tariff) {
    return tariff.tiers.map((t) => ({
        upTo: t.up_to === null || t.up_to === undefined ? "" : String(t.up_to),
        price: String(t.price),
    }));
}

export function tiersToRequest(rows) {
    const last = rows.length - 1;
    return rows.map((r, i) => ({
        up_to: i === last || r.upTo === "" ? null : r.upTo,
        price: r.price === "" ? null : r.price,
    }));
}

export function formatPrice(value) {
    return Number(value).toLocaleString(undefined, { maximumFractionDigits: 4 });
}

function formatBound(value) {
    return Number(value).toLocaleString(undefined, { maximumFractionDigits: 3 });
}

export function isFlat(tariff) {
    return tariff.tiers.length <= 1;
}

export function tierRanges(tariff, unit) {
    if (isFlat(tariff)) {
        return [{ label: "Any volume", price: tariff.tiers[0]?.price }];
    }
    return tariff.tiers.map((t, i) => {
        const lower = i > 0 ? tariff.tiers[i - 1].up_to : null;
        let label;
        if (i === 0) label = `up to ${formatBound(t.up_to)} ${unit}`;
        else if (t.up_to === null) label = `over ${formatBound(lower)} ${unit}`;
        else label = `${formatBound(lower)} – ${formatBound(t.up_to)} ${unit}`;
        return { label, price: t.price };
    });
}

export function tiersSummary(tariff) {
    if (isFlat(tariff)) return `flat · ${formatPrice(tariff.tiers[0]?.price ?? 0)}`;
    const prices = tariff.tiers.map((t) => Number(t.price));
    return `${tariff.tiers.length} tiers · ${formatPrice(Math.min(...prices))} – ${formatPrice(Math.max(...prices))}`;
}
import { useState } from "react";
import { updateRateSource, fetchRates } from "../api/client";

function toDateInput(value) {
    if (!value) return "";
    return value.slice(0, 10);
}

function humanInterval(seconds) {
    const n = Number(seconds);
    if (!n || n <= 0) return "";
    if (n % 86400 === 0) return `${n / 86400} day(s)`;
    if (n % 3600 === 0) return `${n / 3600} hour(s)`;
    if (n % 60 === 0) return `${n / 60} minute(s)`;
    return `${n} second(s)`;
}

function RateSourceCard({ source, providers, onChanged }) {
    const [form, setForm] = useState({
        source: source.source,
        url_template: source.url_template,
        poll_interval_seconds: source.poll_interval_seconds,
        request_timeout_seconds: source.request_timeout_seconds,
        backfill_start: toDateInput(source.backfill_start),
    });
    const [status, setStatus] = useState(null);
    const [fetching, setFetching] = useState(false);
    const [fetchMsg, setFetchMsg] = useState(null);

    const label = source.kind === "fiat" ? "Fiat" : "Crypto";

    function update(field, value) {
        setForm((f) => ({ ...f, [field]: value }));
    }

    async function handleSave() {
        setStatus(null);
        try {
            await updateRateSource(source.kind, {
                source: form.source.trim(),
                url_template: form.url_template.trim(),
                poll_interval_seconds: Number(form.poll_interval_seconds),
                request_timeout_seconds: Number(form.request_timeout_seconds),
                backfill_start: form.backfill_start,
            });
            setStatus({ type: "ok", text: "Saved" });
            onChanged();
        } catch (err) {
            setStatus({ type: "error", text: err.message });
        }
    }

    async function handleFetchNow() {
        setFetching(true);
        setFetchMsg(null);
        try {
            const result = await fetchRates(source.kind);
            const errs = Object.keys(result.errors || {}).length;
            setFetchMsg({
                type: errs ? "error" : "ok",
                text: `Fetched ${result.fetched}, skipped ${result.skipped}` + (errs ? ` — ${errs} error(s)` : ""),
            });
            onChanged();
        } catch (err) {
            setFetchMsg({ type: "error", text: err.message });
        } finally {
            setFetching(false);
        }
    }

    return (
        <div className="card">
            <div className="rate-source-head">
                <h3 className="form-title">{label}</h3>
                <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={handleFetchNow}
                    disabled={fetching}
                >
                    {fetching ? "Fetching..." : "Fetch now"}
                </button>
            </div>

            <div className="form-grid">
                <div className="field">
                    <label>Source</label>
                    <select
                        value={form.source}
                        onChange={(e) => update("source", e.target.value)}
                    >
                        <option value="">Select...</option>
                        {providers.map((p) => (
                            <option key={p} value={p}>{p}</option>
                        ))}
                    </select>
                </div>

                <div className="field">
                    <label>Poll interval (seconds)</label>
                    <input
                        type="number"
                        value={form.poll_interval_seconds}
                        onChange={(e) => update("poll_interval_seconds", e.target.value)} 
                    />
                    <span className="field-hint">{humanInterval(form.poll_interval_seconds)}</span>
                </div>

                <div className="field">
                    <label>Request timeout (seconds)</label>
                    <input 
                        type="number"
                        value={form.request_timeout_seconds}
                        onChange={(e) => update("request_timeout_seconds", e.target.value)}
                    />
                </div>

                <div className="field">
                    <label>Backfill start</label>
                    <input
                        type="date"
                        value={form.backfill_start}
                        onChange={(e) => update("backfill_start", e.target.value)} 
                    />
                </div>
            </div>

            <div className="field" style={{ marginTop: 16 }}>
                <label>URL template</label>
                <input
                    type="text"
                    className="mono"
                    value={form.url_template}
                    onChange={(e) => update("url_template", e.target.value)} 
                />
                <span className="field-hint">
                    Placeholders: {"{base}"} {"{quote}"} {"{date}"}
                </span>
            </div>

            <div className="form-actions">
                {fetchMsg && (
                    <span className={fetchMsg.type === "ok" ? "status-ok" : "status-error"}>
                        {fetchMsg.text}
                    </span>
                )}
                {status && (
                    <span className={status.type === "ok" ? "status-ok" : "status-error"}>
                        {status.text}
                    </span>
                )}
                <button type="button" className="btn btn-primary" onClick={handleSave}>
                    Save
                </button>
            </div>
        </div>
    );
}

export default RateSourceCard;
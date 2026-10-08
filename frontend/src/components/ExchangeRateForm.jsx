import { useState } from "react";
import { createExchangeRate } from "../api/client";

function ExchangeRateForm({ currencies, onCreated }) {
    const [currency, setCurrency] = useState("");
    const [date, setDate] = useState("");
    const [rate, setRate] = useState("");
    const [status, setStatus] = useState(null);

    const options = currencies.filter((c) => c.code !== "USD" && c.is_active);

    async function handleSubmit(event) {
        event.preventDefault();
        setStatus(null);

        if (!currency || !date || !rate.trim()) {
            setStatus({ type: "error", text: "All fields are required" });
            return;
        }

        try {
            await createExchangeRate({
                currency,
                rateAt: `${date}T00:00:00Z`,
                rate: rate.trim(),
            });
            setStatus({ type: "ok", text: `Saved ${currency} rate for ${date}` });
            setCurrency("");
            setDate("");
            setRate("");
            onCreated();
        } catch (err) {
            setStatus({ type: "error", text: err.message });
        }
    }

    return (
        <div className="card">
            <h5 className="form-title">Add a rate</h5>
            <form onSubmit={handleSubmit}>
                <div className="form-inline" style={{ "--fields": 3 }}>
                    <div className="field">
                        <label htmlFor="er-currency">Currency</label>
                        <select 
                            id="er-currency"
                            value={currency}
                            onChange={(e) => setCurrency(e.target.value)}
                        >
                            <option value="">Select...</option>
                            {options.map((c) => (
                                <option key={c.code} value={c.code}>
                                    {c.code} - {c.name}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div className="field">
                        <label htmlFor="er-date">As-of date</label>
                        <input
                            id="er-date"
                            type="date"
                            value={date}
                            onChange={(e) => setDate(e.target.value)} 
                        />
                    </div>

                    <div className="field">
                        <label htmlFor="er-rate">Rate (price in USD)</label>
                        <input
                            id="er-rate"
                            type="text"
                            inputMode="decimal"
                            placeholder="0.00274"
                            value={rate}
                            onChange={(e) => setRate(e.target.value)}
                        />
                    </div>

                    <button type="submit" className="btn btn-primary">Save rate</button>
                </div>
            </form>

            {status && (
                <p className={status.type === "ok" ? "status-ok" : "status-error"}>
                    {status.text}
                </p>
            )}
        </div>
    );
}

export default ExchangeRateForm;
import { useState } from "react";

function formatDate(isoString) {
    return new Date(isoString).toLocaleDateString();
}

function IncomeRow({ income, companiesById }) {
    const [expanded, setExpanded] = useState(false);

    const hasDetails = income.note || income.transaction_ref;

    return (
        <li className="income-item">
            <div
                className="row-income"
                onClick={() => hasDetails && setExpanded(!expanded)}
                style={{ cursor: hasDetails ? "pointer" : "default" }}
            >
                <span className="row-meta">{formatDate(income.occurred_at)}</span>
                <span className="row-key mono">{income.amount} {income.currency}</span>
                <span>
                    {companiesById[income.payer_id] || income.payer_id}
                    {" → "}
                    {companiesById[income.beneficiary_id] || income.beneficiary_id}
                </span>
                <span className="badge">{income.payment_type}</span>
                <span className="row-meta">
                    {hasDetails ? (expanded ? "▲ details" : "▼ details") : ""}
                </span>
            </div>

            {expanded && (
                <div className="income-details">
                    {income.note && (
                        <div><span className="detail-label">Note:</span> {income.note}</div>
                    )}
                    {income.transaction_ref && (
                        <div>
                            <span className="detail-label">Ref:</span> {" "}
                            {income.transaction_ref.startsWith("http") ? (
                                <a href={income.transaction_ref} target="_blank" rel="noreferrer">
                                    {income.transaction_ref}
                                </a>
                            ) : (
                                income.transaction_ref
                            )}
                        </div>
                    )}
                </div>
            )}
        </li>
    );
}

export default IncomeRow;
function trimZeros(value) {
    const s = String(value);
    if (!s.includes(".")) return s;
    return s.replace(/0+$/, "").replace(/\.$/, "");
}

function ExchangeRateRow({ rate }) {
    return (
        <li className="row-exchange-rate">
            <span className="row-key">{rate.currency}</span>
            <span className="rate-value">{trimZeros(rate.rate)}</span>
            <span className="row-meta mono">{rate.rate_at}</span>
            <span className="badge">{rate.source}</span>
        </li>
    );
}

export default ExchangeRateRow;
import CurrencyRow from "./CurrencyRow";

function CurrenciesList({ currencies, onSave, onDeactivate }) {
    return (
        <ul className="list">
            {currencies.map((c) => (
                <CurrencyRow
                    key={c.code}
                    currency={c}
                    onSave={onSave}
                    onDeactivate={onDeactivate}
                />
            ))}
        </ul>
    );
}

export default CurrenciesList;
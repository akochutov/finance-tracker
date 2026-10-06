import { MODE_WHOLE, MODE_PROGRESSIVE, MODE_LABELS } from "./tariffUtils";

function TierEditor({ tiers, onChange, mode, onModeChange, unit, currency }) {
    const last = tiers.length - 1;

    function setRow(i, field, value) {
        onChange(tiers.map((t, j) => (j === i ? { ...t, [field]: value } : t)));
    }

    function addTier() {
        onChange([...tiers, { upTo: "", price: "" }]);
    }

    function removeTier(i) {
        onChange(tiers.filter((_, j) => j !== i));
    }

    return (
        <div className="tf-editor">
            <div className="tf-editor-tiers">
                <div className="tf-tier-row tf-tier-head">
                    <span>Tier</span>
                    <span>Up to{unit ? `, ${unit}` : ""}</span>
                    <span>Price{currency ? `, ${currency}` : ""}</span>
                    <span></span>
                </div>

                {tiers.map((t, i) => (
                    <div key={i} className="tf-tier-row">
                        <span className="row-meta">{i + 1}</span>
                        {i === last ? (
                            <span className="tf-and-above">{tiers.length === 1 ? "any volume" : "and above"}</span>
                        ) : (
                            <input
                                className="input mono"
                                type="number"
                                step="any"
                                min="0"
                                value={t.upTo}
                                onChange={(e) => setRow(i, "upTo", e.target.value)}
                            />
                        )}
                        <input
                            className="input mono"
                            type="number"
                            step="any"
                            min="0"
                            value={t.price}
                            onChange={(e) => setRow(i, "price", e.target.value)}
                        />
                        {tiers.length > 1 ? (
                            <button type="button" className="tf-remove" aria-label={`Remove tier ${i + 1}`} onClick={() => removeTier(i)}>×</button>
                        ) : (
                            <span></span>
                        )}
                    </div>
                ))}

                <button type="button" className="btn btn-secondary btn-sm tf-add" onClick={addTier}>+ Tier</button>
            </div>

            {tiers.length > 1 && (
                <div className="tf-modes">
                    <div className="mt-info-label">How tiers apply</div>
                    <label className="tf-mode">
                        <input type="radio" checked={mode !== MODE_PROGRESSIVE} onChange={() => onModeChange(MODE_WHOLE)} />
                        <span>
                            <strong>{MODE_LABELS[MODE_WHOLE]}</strong>
                            <span className="tf-mode-hint">The whole month at the price of the tier its volume falls into.</span>
                        </span>
                    </label>
                    <label className="tf-mode">
                        <input type="radio" checked={mode === MODE_PROGRESSIVE} onChange={() => onModeChange(MODE_PROGRESSIVE)} />
                        <span>
                            <strong>{MODE_LABELS[MODE_PROGRESSIVE]}</strong>
                            <span className="tf-mode-hint">Each part of the volume at its own tier's price.</span>
                        </span>
                    </label>
                    <div className="tf-mode-hint tf-mode-note">
                        The volume is the account's monthly total: day and night together pick the tier.
                    </div>
                </div>
            )}
        </div>
    );
}

export default TierEditor;
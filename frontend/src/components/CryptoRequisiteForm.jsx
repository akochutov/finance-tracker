import { useState } from "react";
import { createCryptoRequisite } from "../api/client";

function CryptoRequisiteForm({ companyId, onCreated }) {
    const [network, setNetwork] = useState("");
    const [walletAddress, setWalletAddress] = useState("");
    const [note, setNote] = useState("");
    const [error, setError] = useState(null);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await createCryptoRequisite(companyId, {
                network: network,
                wallet_address: walletAddress,
                note: note || null,
            });
            setNetwork("");
            setWalletAddress("");
            setNote("");
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form className="card" onSubmit={handleSubmit}>
            <h5 className="form-title">Add crypto requisite</h5>
            {error && <div className="error">{error}</div>}
            <div className="form-grid">
                <div className="field">
                    <label>Network</label>
                    <input className="input" placeholder="TRC20" value={network} onChange={(e) => setNetwork(e.target.value)} />
                </div>
                <div className="field">
                    <label>Wallet address</label>
                    <input className="input" value={walletAddress} onChange={(e) => setWalletAddress(e.target.value)} />
                </div>
                <div className="field">
                    <label>Note</label>
                    <input className="input" placeholder="Optional comment" value={note} onChange={(e) => setNote(e.target.value)} />
                </div>
            </div>
            <div className="form-actions">
                <button type="submit" className="btn btn-primary">Add</button>
            </div>
        </form>
    );
}

export default CryptoRequisiteForm;
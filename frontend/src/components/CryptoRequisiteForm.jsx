import { useState } from "react";
import { createCryptoRequisite } from "../api/client";

function CryptoRequisiteForm({ companyId, onCreated }) {
    const [network, setNetwork] = useState("");
    const [walletAddress, setWalletAddress] = useState("");
    const [error, setError] = useState(null);

    async function handleSubmit(e) {
        e.preventDefault();
        setError(null);
        try {
            await createCryptoRequisite(companyId, {
                network: network,
                wallet_address: walletAddress,
            });
            setNetwork("");
            setWalletAddress("");
            onCreated();
        } catch (err) {
            setError(err.message);
        }
    }

    return (
        <form onSubmit={handleSubmit}>
            <h4>Add crypto requisite</h4>
            {error && <div style={{ color: "red" }}>{error}</div>}
            <input
                placeholder="Network (TRC20)"
                value={network}
                onChange={(e) => setNetwork(e.target.value)}
            />
            <input
                placeholder="Wallet address"
                value={walletAddress}
                onChange={(e) => setWalletAddress(e.target.value)}
            />
            <button type="submit">Add</button>
        </form>
    );
}

export default CryptoRequisiteForm;
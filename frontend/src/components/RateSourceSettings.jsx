import { useState, useEffect, useCallback } from "react";
import { getRateSources, getRateProviders } from "../api/client";
import RateSourceCard from "./RateSourceCard";

function RateSourceSettings() {
    const [sources, setSources] = useState([]);
    const [providers, setProviders] = useState([]);
    const [error, setError] = useState(null);

    const load = useCallback(() => {
        Promise.all([getRateSources(), getRateProviders()])
            .then(([srcs, provs]) => {
                setSources(srcs);
                setProviders(provs);
            })
            .catch((err) => setError(err.message));
    }, []);

    useEffect(() => {
        load();
    }, [load]);

    return (
        <div className="section">
            <h3>Rate sources</h3>
            {error && <div className="error">{error}</div>}
            {sources.map((s) => (
                <RateSourceCard
                    key={s.kind}
                    source={s}
                    providers={providers}
                    onChanged={load}
                />
            ))}
        </div>
    );
}

export default RateSourceSettings;
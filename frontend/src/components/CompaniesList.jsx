import { useState, useEffect } from "react";
import { getCompanies } from "../api/client";

function CompaniesList() {
    const [companies, setCompanies] = useState([]);
    const [error, setError] = useState(null);

    useEffect(() => {
        getCompanies()
            .then((data) => setCompanies(data))
            .catch((err) => setError(err.message));
    }, [])

    if (error) {
        return <div>Error loading companies: {error}</div>;
    }

    return (
        <div>
            <h2>Companies</h2>
            <ul>
                {companies.map((c) => (
                    <li key={c.id}>
                        {c.id} - {c.name} ({c.tax_id || "-"}, {c.address || "-"} {c.note || "-"})
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default CompaniesList;
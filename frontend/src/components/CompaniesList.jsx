import CompanyRow from "./CompanyRow";

function CompaniesList({ companies, onSave, onDeactivate }) {
    return (
        <ul>
            {companies.map((c) => (
                <CompanyRow 
                    key={c.id}
                    company={c}
                    onSave={onSave}
                    onDeactivate={onDeactivate}
                />
            ))}
        </ul>
    );
}

export default CompaniesList;
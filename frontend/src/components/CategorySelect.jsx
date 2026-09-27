function CategorySelect({ groups, value, onChange, keepIds }) {
    return (
        <select className="input" value={value} onChange={(e) => onChange(e.target.value)}>
            <option value="">- Category -</option>
            {groups.map((g) => {
                const options = g.categories.filter((c) => c.is_active || keepIds.has(c.id));
                if (options.length === 0) {
                    return null;
                }
                return (
                    <optgroup key={g.id} label={g.name}>
                        {options.map((c) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                        ))}
                    </optgroup>
                );
            })}
        </select>
    );
}

export default CategorySelect;
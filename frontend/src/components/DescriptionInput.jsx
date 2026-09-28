import { useState, useMemo } from "react";

const MAX_OPTIONS = 8;

function rankSuggestions(suggestions, query) {
    const q = query.trim().toLowerCase();
    if (q === "") {
        return [];
    }

    const starts = [];
    const contains = [];
    for (const s of suggestions) {
        const pos = s.lower.indexOf(q);
        if (pos === 0) {
            starts.push(s);
        } else if (pos > 0) {
            contains.push(s);
        }
    }

    const byUses = (a, b) => b.uses - a.uses;
    starts.sort(byUses);
    contains.sort(byUses);
    return starts.concat(contains).slice(0, MAX_OPTIONS);
}

function DescriptionInput({ value, onChange, onPick, suggestions, categoriesById }) {
    const [open, setOpen] = useState(false);
    const [active, setActive] = useState(0);

    const options = useMemo(() => rankSuggestions(suggestions, value), [suggestions, value]);
    const visible = open && options.length > 0;
    const current = Math.min(active, options.length - 1);

    function pick(s) {
        onPick(s);
        setOpen(false);
    }

    function handleChange(e) {
        onChange(e.target.value);
        setOpen(true);
        setActive(0);
    }

    function handleKeyDown(e) {
        if (!visible) {
            if (e.key === "ArrowDown" && options.length > 0) {
                e.preventDefault();
                setOpen(true);
                setActive(0);
            }
            return;
        }

        switch (e.key) {
            case "ArrowDown":
                e.preventDefault();
                setActive((current + 1) % options.length);
                break;
            case "ArrowUp":
                e.preventDefault();
                setActive((current - 1 + options.length) % options.length);
                break;
            case "Enter":
                e.preventDefault();
                pick(options[current]);
                break;
            case "Escape":
                e.preventDefault();
                setOpen(false);
                break;
            case "Tab":
                setOpen(false);
                break;
            default:
                break;
        }
    }

    return (
        <div className="suggest">
            <input
                className="input"
                value={value}
                onChange={handleChange}
                onKeyDown={handleKeyDown}
                onBlur={() => setOpen(false)}
                autoComplete="off"
                role="combobox"
                aria-expanded={visible}
                aria-autocomplete="list"
            />
            {visible && (
                <ul className="suggest-list" role="listbox">
                    {options.map((s, i) => (
                        <li
                            key={s.description}
                            role="option"
                            aria-selected={i === current}
                            className={i === current ? "suggest-option active" : "suggest-option"}
                            onMouseDown={(e) => {
                                e.preventDefault();
                                pick(s);
                            }}
                            onMouseEnter={() => setActive(i)}
                        >
                            <span className="suggest-text">{s.description}</span>
                            <span className="suggest-meta">
                                {s.category_id ? categoriesById[s.category_id] || "" : "no category"}
                            </span>
                            <span className="suggest-uses mono">{s.uses}</span>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
}

export default DescriptionInput;
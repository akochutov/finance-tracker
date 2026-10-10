import { useState } from "react";
import { formatISODate, todayISO } from "./expenseUtils";
import { isDayNight, zonesOf, zoneLabel, meterTitle, REGISTERS_SINGLE, REGISTERS_DAY_NIGHT } from "./meterUtils";

function PaymentCategorySelect({ groups, value, onChange }) {
    return (
        <select className="input" value={value || ""} onChange={(e) => onChange(e.target.value)}>
            <option value="">Payments: not tracked</option>
            {groups.map((g) => {
                const cats = (g.categories || []).filter((c) => c.is_active || c.id === value);
                if (cats.length === 0) return null;
                return (
                    <optgroup key={g.id} label={g.name}>
                        {cats.map((c) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                        ))}
                    </optgroup>
                );
            })}
        </select>
    );
}

function categoryName(groups, id) {
    if (!id) return null;
    for (const g of groups) {
        const c = (g.categories || []).find((x) => x.id === id);
        if (c) return `${g.name} → ${c.name}`;
    }
    return null;
}

function MetersTree({ addresses, accounts, meters, serviceTypes, groups, selectedId, onSelect, actions }) {
    const [newAddress, setNewAddress] = useState("");

    async function addAddress(e) {
        e.preventDefault();
        if (await actions.createAddress(newAddress)) setNewAddress("");
    }

    const sorted = [...addresses].sort((a, b) => Number(b.is_active) - Number(a.is_active));

    return (
        <section className="xd-panel mt-tree">
            <form className="mt-inline-form" onSubmit={addAddress}>
                <input
                    className="input"
                    placeholder="New address"
                    value={newAddress}
                    onChange={(e) => setNewAddress(e.target.value)}
                />
                <button type="submit" className="btn btn-primary btn-sm">Add</button>
            </form>

            {sorted.map((a) => (
                <AddressNode
                    key={a.id}
                    address={a}
                    accounts={accounts.filter((acc) => acc.address_id === a.id)}
                    meters={meters}
                    serviceTypes={serviceTypes}
                    groups={groups}
                    selectedId={selectedId}
                    onSelect={onSelect}
                    actions={actions}
                />
            ))}
            {addresses.length === 0 && <p className="xd-note">No addresses yet.</p>}
        </section>
    );
}

function AddressNode({ address, accounts, meters, serviceTypes, groups, selectedId, onSelect, actions }) {
    const [expanded, setExpanded] = useState(address.is_active);
    const [editing, setEditing] = useState(false);
    const [name, setName] = useState(address.address);
    const [service, setService] = useState(serviceTypes[0]?.code || "");
    const [number, setNumber] = useState("");
    const [categoryId, setCategoryId] = useState("");

    const meterCount = meters.filter((m) => accounts.some((acc) => acc.id === m.account_id)).length;

    async function save(e) {
        e.preventDefault();
        if (await actions.updateAddress(address.id, name)) setEditing(false);
    }

    async function addAccount(e) {
        e.preventDefault();
        if (await actions.createAccount(address.id, service, number, categoryId)) {
            setNumber("");
            setCategoryId("");
        }
    }

    return (
        <div className={address.is_active ? "mt-address" : "mt-address inactive"}>
            {editing ? (
                <form className="mt-inline-form" onSubmit={save}>
                    <input className="input" value={name} onChange={(e) => setName(e.target.value)} />
                    <button type="submit" className="btn btn-primary btn-sm">Save</button>
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => setEditing(false)}>Cancel</button>
                </form>
            ) : (
                <div className="mt-address-head">
                    <button type="button" className="mt-address-name" onClick={() => setExpanded(!expanded)}>
                        <span className="group-toggle-icon">{expanded ? "▼" : "▶"}</span>
                        <span className="mt-address-text">
                            <span className="mt-address-title">{address.address}</span>
                            <span className="mt-meta">
                                {address.is_active ? "" : "inactive · "}
                                {accounts.length} {accounts.length === 1 ? "account" : "accounts"} · {meterCount} {meterCount === 1 ? "meter" : "meters"}
                            </span>
                        </span>
                    </button>
                    <div className="mt-actions">
                        {address.is_active ? (
                            <>
                                <button className="btn btn-secondary btn-sm" onClick={() => { setName(address.address); setEditing(true); }}>Edit</button>
                                <button className="btn btn-ghost btn-sm" onClick={() => actions.setAddressActive(address.id, false)}>Deactivate</button>
                            </>
                        ) : (
                            <button className="btn btn-secondary btn-sm" onClick={() => actions.setAddressActive(address.id, true)}>Activate</button>
                        )}
                    </div>
                </div>
            )}

            {expanded && (
                <div className="mt-accounts">
                    {accounts.map((acc) => (
                        <AccountNode
                            key={acc.id}
                            account={acc}
                            serviceType={serviceTypes.find((t) => t.code === acc.service)}
                            groups={groups}
                            meters={meters.filter((m) => m.account_id === acc.id)}
                            selectedId={selectedId}
                            onSelect={onSelect}
                            actions={actions}
                        />
                    ))}

                    {address.is_active && (
                        <form className="mt-inline-form mt-subform" onSubmit={addAccount}>
                            <select className="input" value={service} onChange={(e) => setService(e.target.value)}>
                                {serviceTypes.map((t) => (
                                    <option key={t.code} value={t.code}>{t.name}</option>
                                ))}
                            </select>
                            <input
                                className="input"
                                placeholder="Account number"
                                value={number}
                                onChange={(e) => setNumber(e.target.value)}
                            />
                            <PaymentCategorySelect groups={groups} value={categoryId} onChange={setCategoryId} />
                            <button type="submit" className="btn btn-secondary btn-sm">+ Account</button>
                        </form>
                    )}
                </div>
            )}
        </div>
    );
}

function AccountNode({ account, serviceType, groups, meters, selectedId, onSelect, actions }) {
    const [editing, setEditing] = useState(false);
    const [number, setNumber] = useState(account.number);
    const [categoryId, setCategoryId] = useState(account.expense_category_id || "");
    const [adding, setAdding] = useState(false);
    const [serial, setSerial] = useState("");
    const [name, setName] = useState("");
    const [installedOn, setInstalledOn] = useState(todayISO());
    const [registers, setRegisters] = useState(REGISTERS_SINGLE);
    const [initialValues, setInitialValues] = useState({});
    const [initialOn, setInitialOn] = useState("");

    async function saveAccount(e) {
        e.preventDefault();
        if (await actions.updateAccount(account.id, number, categoryId)) setEditing(false);
    }

    function startEdit() {
        setNumber(account.number);
        setCategoryId(account.expense_category_id || "");
        setEditing(true);
    }

    const category = categoryName(groups, account.expense_category_id);

    const newZones = zonesOf({ registers });

    async function addMeter(e) {
        e.preventDefault();
        const values = {};
        for (const z of newZones) {
            if (initialValues[z] !== undefined && initialValues[z] !== "") values[z] = initialValues[z];
        }
        const ok = await actions.createMeter({
            account_id: account.id,
            serial,
            name,
            registers,
            installed_on: installedOn,
            removed_on: null,
            initial_on: initialOn || null,
            initial_values: values,
        });
        if (ok) {
            setAdding(false);
            setSerial("");
            setName("");
            setRegisters(REGISTERS_SINGLE);
            setInitialValues({});
            setInitialOn("");
        }
    }

    return (
        <div className={account.is_active ? "mt-account" : "mt-account inactive"}>
            {editing ? (
                <form className="mt-inline-form mt-subform" onSubmit={saveAccount}>
                    <input className="input" value={number} onChange={(e) => setNumber(e.target.value)} />
                    <PaymentCategorySelect groups={groups} value={categoryId} onChange={setCategoryId} />
                    <button type="submit" className="btn btn-primary btn-sm">Save</button>
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => setEditing(false)}>Cancel</button>
                </form>
            ) : (
                <div className="mt-account-head">
                    <span>
                        <strong>{serviceType ? serviceType.name : account.service}</strong>{" "}
                        <span className="mono row-meta">· {account.number}</span>
                        {!account.is_active && <span className="row-meta"> · inactive</span>}
                    </span>
                    <span className="mt-actions">
                        {account.is_active ? (
                            <>
                                <button type="button" className="mt-link" onClick={() => setAdding(!adding)}>+ meter</button>
                                <button type="button" className="mt-link" onClick={startEdit}>edit</button>
                                <button type="button" className="mt-link mt-link-danger" onClick={() => actions.setAccountActive(account.id, false)}>deactivate</button>
                            </>
                        ) : (
                            <button type="button" className="mt-link" onClick={() => actions.setAccountActive(account.id, true)}>activate</button>
                        )}
                    </span>
                </div>
            )}

            {!editing && (
                <div className="mt-category">
                    {category ? `→ ${category}` : "payments not tracked"}
                </div>
            )}

            <div className="mt-meters">
                {meters.map((m) => {
                    const classes = ["mt-meter"];
                    if (m.id === selectedId) classes.push("selected");
                    if (m.removed_on) classes.push("removed");
                    return (
                        <button type="button" key={m.id} className={classes.join(" ")} onClick={() => onSelect(m.id)}>
                            <span className={m.name ? "" : "mono"}>{meterTitle(m)}</span>
                            {isDayNight(m) && <span className="mt-zone-badge dual">day/night</span>}
                            <span className="mt-since">
                                {m.removed_on
                                    ? `removed ${formatISODate(m.removed_on)}`
                                    : `since ${formatISODate(m.installed_on)}`}
                            </span>
                        </button>
                    );
                })}
            </div>

            {adding && (
                <form className="mt-meter-form" onSubmit={addMeter}>
                    <div className="field">
                        <label>Name</label>
                        <input className="input" placeholder="Optional, e.g. Д12-39-Ванна" value={name} onChange={(e) => setName(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Serial</label>
                        <input className="input" value={serial} onChange={(e) => setSerial(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Installed</label>
                        <input className="input" type="date" value={installedOn} onChange={(e) => setInstalledOn(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Registers</label>
                        <select className="input" value={registers} onChange={(e) => setRegisters(e.target.value)}>
                            <option value={REGISTERS_SINGLE}>Single</option>
                            <option value={REGISTERS_DAY_NIGHT}>Day / night</option>
                        </select>
                    </div>
                    <div className="field">
                        <label>Initial values on</label>
                        <input className="input" type="date" value={initialOn} onChange={(e) => setInitialOn(e.target.value)} />
                    </div>
                    {newZones.map((z) => (
                        <div className="field" key={z}>
                            <label>
                                {zoneLabel(z) ? `Initial ${zoneLabel(z).toLowerCase()}` : "Initial value"}
                                {serviceType ? `, ${serviceType.unit}` : ""}
                            </label>
                            <input
                                className="input mono"
                                type="number"
                                step="any"
                                min="0"
                                value={initialValues[z] ?? ""}
                                onChange={(e) => setInitialValues({ ...initialValues, [z]: e.target.value })}
                            />
                        </div>
                    ))}
                    <div className="mt-hint mt-meter-form-hint">
                        What the meter shows when tracking starts — not counted as consumption.
                        A day/night meter needs both registers. Leave the date empty to use the
                        installation day; set it to the move-in day otherwise.
                    </div>
                    <div className="mt-meter-form-actions">
                        <button type="submit" className="btn btn-primary btn-sm">Add</button>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={() => setAdding(false)}>Cancel</button>
                    </div>
                </form>
            )}
        </div>
    );
}

export default MetersTree;
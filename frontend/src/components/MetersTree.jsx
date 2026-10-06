import { useState } from "react";
import { formatISODate, todayISO } from "./expenseUtils";
import { isDayNight, ZONES_SINGLE, ZONES_DAY_NIGHT } from "./meterUtils";

function MetersTree({ addresses, accounts, meters, serviceTypes, selectedId, onSelect, actions }) {
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
                    selectedId={selectedId}
                    onSelect={onSelect}
                    actions={actions}
                />
            ))}
            {addresses.length === 0 && <p className="xd-note">No addresses yet.</p>}
        </section>
    );
}

function AddressNode({ address, accounts, meters, serviceTypes, selectedId, onSelect, actions }) {
    const [expanded, setExpanded] = useState(address.is_active);
    const [editing, setEditing] = useState(false);
    const [name, setName] = useState(address.address);
    const [service, setService] = useState(serviceTypes[0]?.code || "");
    const [number, setNumber] = useState("");
    const [zones, setZones] = useState(ZONES_SINGLE);

    const meterCount = meters.filter((m) => accounts.some((acc) => acc.id === m.account_id)).length;

    async function save(e) {
        e.preventDefault();
        if (await actions.updateAddress(address.id, name)) setEditing(false);
    }

    async function addAccount(e) {
        e.preventDefault();
        if (await actions.createAccount(address.id, service, number, zones)) {
            setNumber("");
            setZones(ZONES_SINGLE);
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
                            {/* Zones are a term of the contract: chosen once, never changed */}
                            <select className="input" value={zones} onChange={(e) => setZones(e.target.value)}>
                                <option value={ZONES_SINGLE}>Single zone</option>
                                <option value={ZONES_DAY_NIGHT}>Day / night</option>
                            </select>
                            <button type="submit" className="btn btn-secondary btn-sm">+ Account</button>
                        </form>
                    )}
                </div>
            )}
        </div>
    );
}

function AccountNode({ account, serviceType, meters, selectedId, onSelect, actions }) {
    const [editing, setEditing] = useState(false);
    const [number, setNumber] = useState(account.number);
    const [adding, setAdding] = useState(false);
    const [serial, setSerial] = useState("");
    const [installedOn, setInstalledOn] = useState(todayISO());
    const [initialValue, setInitialValue] = useState("");
    const [initialOn, setInitialOn] = useState("");

    async function saveNumber(e) {
        e.preventDefault();
        if (await actions.updateAccount(account.id, number)) setEditing(false);
    }

    async function addMeter(e) {
        e.preventDefault();
        const ok = await actions.createMeter({
            account_id: account.id,
            serial,
            installed_on: installedOn,
            removed_on: null,
            initial_on: initialOn || null,
            initial_value: initialValue === "" ? null : initialValue,
        });
        if (ok) {
            setAdding(false);
            setSerial("");
            setInitialValue("");
            setInitialOn("");
        }
    }

    return (
        <div className={account.is_active ? "mt-account" : "mt-account inactive"}>
            {editing ? (
                <form className="mt-inline-form" onSubmit={saveNumber}>
                    <input className="input" value={number} onChange={(e) => setNumber(e.target.value)} />
                    <button type="submit" className="btn btn-primary btn-sm">Save</button>
                    <button type="button" className="btn btn-secondary btn-sm" onClick={() => setEditing(false)}>Cancel</button>
                </form>
            ) : (
                <div className="mt-account-head">
                    <span>
                        <strong>{serviceType ? serviceType.name : account.service}</strong>{" "}
                        <span className="mono row-meta">· {account.number}</span>{" "}
                        <span className={isDayNight(account) ? "mt-zone-badge dual" : "mt-zone-badge"}>
                            {isDayNight(account) ? "day/night" : "single"}
                        </span>
                        {!account.is_active && <span className="row-meta"> · inactive</span>}
                    </span>
                    <span className="mt-actions">
                        {account.is_active ? (
                            <>
                                <button type="button" className="mt-link" onClick={() => setAdding(!adding)}>+ meter</button>
                                <button type="button" className="mt-link" onClick={() => { setNumber(account.number); setEditing(true); }}>edit</button>
                                <button type="button" className="mt-link mt-link-danger" onClick={() => actions.setAccountActive(account.id, false)}>deactivate</button>
                            </>
                        ) : (
                            <button type="button" className="mt-link" onClick={() => actions.setAccountActive(account.id, true)}>activate</button>
                        )}
                    </span>
                </div>
            )}

            <div className="mt-meters">
                {meters.map((m) => {
                    const classes = ["mt-meter"];
                    if (m.id === selectedId) classes.push("selected");
                    if (m.removed_on) classes.push("removed");
                    return (
                        <button type="button" key={m.id} className={classes.join(" ")} onClick={() => onSelect(m.id)}>
                            <span className="mono">{m.serial}</span>
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
                        <label>Serial</label>
                        <input className="input" value={serial} onChange={(e) => setSerial(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Installed</label>
                        <input className="input" type="date" value={installedOn} onChange={(e) => setInstalledOn(e.target.value)} />
                    </div>
                    <div className="field">
                        <label>Initial value{serviceType ? `, ${serviceType.unit}` : ""}</label>
                        <input
                            className="input mono"
                            type="number"
                            step="any"
                            min="0"
                            value={initialValue}
                            onChange={(e) => setInitialValue(e.target.value)}
                        />
                    </div>
                    <div className="field">
                        <label>Initial value on</label>
                        <input className="input" type="date" value={initialOn} onChange={(e) => setInitialOn(e.target.value)} />
                    </div>
                    <div className="mt-hint mt-meter-form-hint">
                        The value on the meter when tracking starts — not counted as consumption.
                        Leave the date empty to use the installation day; set it to the move-in day otherwise.
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
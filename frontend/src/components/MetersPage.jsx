import { useState, useEffect } from "react";
import {
    getServiceTypes, getAddresses, getUtilityAccounts, getMeters,
    createAddress, updateAddress, deactivateAddress, activateAddress,
    createUtilityAccount, updateUtilityAccount, deactivateUtilityAccount, activateUtilityAccount,
    createMeter, updateMeter, deleteMeter,
} from "../api/client";
import MetersTree from "./MetersTree";
import MeterPanel from "./MeterPanel";

function MetersPage() {
    const [serviceTypes, setServiceTypes] = useState([]);
    const [addresses, setAddresses] = useState([]);
    const [accounts, setAccounts] = useState([]);
    const [meters, setMeters] = useState([]);
    const [selectedId, setSelectedId] = useState(null);
    const [loaded, setLoaded] = useState(false);
    const [error, setError] = useState(null);

    async function loadAll() {
        try {
            const [types, addrs, accs, mts] = await Promise.all([
                getServiceTypes(), getAddresses(), getUtilityAccounts(), getMeters(),
            ]);
            setServiceTypes(types);
            setAddresses(addrs);
            setAccounts(accs);
            setMeters(mts);
            setLoaded(true);
        } catch (err) {
            setError(err.message);
        }
    }

    useEffect(() => {
        loadAll();
    }, []);

    async function run(action) {
        setError(null);
        try {
            const result = await action();
            await loadAll();
            return result ?? true;
        } catch (err) {
            setError(err.message);
            window.scrollTo({ top: 0, behavior: "smooth" });
            return false;
        }
    }

    const actions = {
        createAddress: (address) => run(() => createAddress(address)),
        updateAddress: (id, address) => run(() => updateAddress(id, address)),
        setAddressActive: (id, active) => run(() => (active ? activateAddress(id) : deactivateAddress(id))),

        createAccount: (addressId, service, number, zones) =>
            run(() => createUtilityAccount({ address_id: addressId, service, number, zones })),
        updateAccount: (id, number) => run(() => updateUtilityAccount(id, number)),
        setAccountActive: (id, active) =>
            run(() => (active ? activateUtilityAccount(id) : deactivateUtilityAccount(id))),

        createMeter: async (fields) => {
            const created = await run(() => createMeter(fields));
            if (created && created.id) setSelectedId(created.id);
            return Boolean(created);
        },
        updateMeter: (id, fields) => run(() => updateMeter(id, fields)),
        deleteMeter: async (id) => {
            const ok = await run(() => deleteMeter(id));
            if (ok) setSelectedId(null);
            return ok;
        },
    };

    const selected = meters.find((m) => m.id === selectedId) || meters[0] || null;
    const account = selected ? accounts.find((a) => a.id === selected.account_id) : null;
    const serviceType = account ? serviceTypes.find((t) => t.code === account.service) : null;

    return (
        <div className="xd-page">
            <div className="page-header">
                <h1>Meters</h1>
                <p>Addresses, accounts and meters.</p>
            </div>

            {error && <div className="error">{error}</div>}
            {!loaded && !error && <div className="loading">Loading…</div>}

            {loaded && (
                <div className="mt-layout">
                    <MetersTree
                        addresses={addresses}
                        accounts={accounts}
                        meters={meters}
                        serviceTypes={serviceTypes}
                        selectedId={selected ? selected.id : null}
                        onSelect={setSelectedId}
                        actions={actions}
                    />
                    {selected && account ? (
                        <MeterPanel
                            key={selected.id}
                            meter={selected}
                            account={account}
                            serviceType={serviceType}
                            actions={actions}
                        />
                    ) : (
                        <section className="xd-panel">
                            <p className="xd-note">Add an address, a utility account and a meter; readings are entered on the Readings page.</p>
                        </section>
                    )}
                </div>
            )}
        </div>
    );
}

export default MetersPage;
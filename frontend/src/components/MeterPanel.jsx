import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { getReadings } from "../api/client";
import { formatISODate } from "./expenseUtils";
import { isDayNight, zonesOf, zoneLabel, groupByDate, formatReading } from "./meterUtils";

function MeterPanel({ meter, account, serviceType, actions }) {
    const unit = serviceType ? serviceType.unit : "";

    const [readings, setReadings] = useState([]);
    const [editing, setEditing] = useState(false);
    const [serial, setSerial] = useState(meter.serial);
    const [installedOn, setInstalledOn] = useState(meter.installed_on.slice(0, 10));
    const [removedOn, setRemovedOn] = useState(meter.removed_on ? meter.removed_on.slice(0, 10) : "");

    useEffect(() => {
        let ignore = false;
        getReadings(meter.id)
            .then((data) => {
                if (!ignore) setReadings(data);
            })
            .catch(() => {
                if (!ignore) setReadings([]);
            });
        return () => {
            ignore = true;
        };
    }, [meter.id, meter.updated_at]);

    async function saveMeter(e) {
        e.preventDefault();
        const ok = await actions.updateMeter(meter.id, {
            serial,
            installed_on: installedOn,
            removed_on: removedOn || null,
        });
        if (ok) setEditing(false);
    }

    async function removeMeter() {
        if (window.confirm(`Delete meter ${meter.serial} with all its readings? This cannot be undone.`)) {
            await actions.deleteMeter(meter.id);
        }
    }

    const zones = zonesOf(meter);
    const rounds = groupByDate(readings);
    const initial = rounds.find((g) => g.is_initial) || null;
    const latest = rounds[0] || null;

    function valuesOf(round) {
        if (!round) return "—";
        return zones
            .map((z) => {
                const r = round.byZone[z];
                const value = r ? formatReading(r.value) : "—";
                return zoneLabel(z) ? `${zoneLabel(z).toLowerCase()} ${value}` : value;
            })
            .join(" · ") + ` ${unit}`;
    }

    return (
        <section className="xd-panel">
            <div className="mt-panel-head">
                <div>
                    <div className="mt-caption">
                        {serviceType ? serviceType.name : account.service} · account {account.number}
                    </div>
                    <div className="mt-title">Meter <span className="mono">{meter.serial}</span></div>
                    <div className="mt-pills">
                        <span className="mt-pill accent">{isDayNight(meter) ? "Day / night registers" : "Single register"}</span>
                        <span className="mt-pill">Installed {formatISODate(meter.installed_on)}</span>
                        {meter.removed_on && <span className="mt-pill">Removed {formatISODate(meter.removed_on)}</span>}
                        <span className="mt-pill">Unit: {unit}</span>
                    </div>
                </div>
                {!editing && (
                    <div className="mt-actions">
                        <button className="btn btn-secondary btn-sm" onClick={() => setEditing(true)}>Edit</button>
                        <button className="btn btn-ghost btn-sm" onClick={removeMeter}>Delete</button>
                    </div>
                )}
            </div>

            {editing && (
                <form className="mt-section" onSubmit={saveMeter}>
                    <div className="mt-section-title">Edit meter</div>
                    <div className="mt-reading-form">
                        <div className="field">
                            <label>Serial</label>
                            <input className="input" value={serial} onChange={(e) => setSerial(e.target.value)} />
                        </div>
                        <div className="field">
                            <label>Installed</label>
                            <input className="input" type="date" value={installedOn} onChange={(e) => setInstalledOn(e.target.value)} />
                        </div>
                        <div className="field">
                            <label>Removed</label>
                            <input className="input" type="date" value={removedOn} onChange={(e) => setRemovedOn(e.target.value)} />
                        </div>
                        <button type="submit" className="btn btn-primary btn-sm">Save</button>
                        <button type="button" className="btn btn-secondary btn-sm" onClick={() => setEditing(false)}>Cancel</button>
                    </div>
                    <div className="mt-hint">Leave “Removed” empty while the meter is in place. All readings, the initial one included, must stay within these dates.</div>
                </form>
            )}

            <div className="mt-info-grid">
                <div>
                    <div className="mt-info-label">Initial reading</div>
                    <div className="mt-info-value mono">{initial ? valuesOf(initial) : "—"}</div>
                    <div className="mt-info-sub">{initial ? `${formatISODate(initial.taken_on)} · not counted` : ""}</div>
                </div>
                <div>
                    <div className="mt-info-label">Last reading</div>
                    <div className="mt-info-value mono">{latest ? valuesOf(latest) : "—"}</div>
                    <div className="mt-info-sub">{latest ? formatISODate(latest.taken_on) : ""}</div>
                </div>
                <div>
                    <div className="mt-info-label">Readings</div>
                    <div className="mt-info-value">{rounds.length}</div>
                    <div className="mt-info-sub">
                        <Link to={`/readings?meter=${meter.id}`}>Open in Readings →</Link>
                    </div>
                </div>
            </div>
        </section>
    );
}

export default MeterPanel;
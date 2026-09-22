import { useState, useEffect, useRef, useCallback } from "react";
import { startBackfill, getBackfillStatus } from "../api/client";

function BackfillPanel({ kind }) {
    const [job, setJob] = useState(null);
    const [error, setError] = useState(null);
    const timer = useRef(null);

    const poll = useCallback(async () => {
        try {
            const status = await getBackfillStatus(kind);
            setJob(status);
            return status.status;
        } catch (err) {
            if (!String(err.message).includes("404")) setError(err.message);
            return null;
        }
    }, [kind]);

    useEffect(() => {
        let cancelled = false;
        poll().then((s) => {
            if (!cancelled && s === "running") startPolling();
        });
        return () => {
            cancelled = true;
            stopPolling();
        };
    }, [poll]);

    function startPolling() {
        stopPolling();
        timer.current = setInterval(async () => {
            const s = await poll();
            if (s !== "running") stopPolling();
        }, 1000);
    }

    function stopPolling() {
        if (timer.current) {
            clearInterval(timer.current);
            timer.current = null;
        }
    }

    async function handleStart() {
        setError(null);
        try {
            const started = await startBackfill(kind);
            setJob(started);
            startPolling();
        } catch (err) {
            setError(err.message);
        }
    }

    const running = job?.status === "running";
    const pct = 
        job && job.total_days > 0
            ? Math.min(100, Math.round((job.done_days / job.total_days) * 100))
            : 0;
    
    return (
        <div className="backfill">
            <div className="backfill-head">
                <span className="backfill-title">Backfill history</span>
                <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={handleStart}
                    disabled={running}
                >
                    {running ? "Running..." : "Start backfill"}
                </button>
            </div>

            {error && <div className="error" style={{ marginTop: 10 }}>{error}</div>}

            {job && (
                <div className="backfill-status">
                    <div className="backfill-bar">
                        <div className="backfill-bar-fill" style={{ width: `${pct}%` }}/>
                    </div>
                    <div className="backfill-line">
                        {running ? (
                            <>
                                Fetching <strong>{job.current_currency}</strong> @{" "}
                                <span className="mono">{job.current_date}</span>
                            </>
                        ) : job.status === "done" ? (
                            <>Done.</>
                        ) : (
                            <>Failed: {job.error}</>
                        )}
                        <span className="backfill-counts">
                            {job.done_days}/{job.total_days} days · stored {job.stored}
                            {job.errors > 0 ? ` · ${job.errors} errors` : ""}
                        </span>
                    </div>
                </div>
            )}
        </div>
    );
}

export default BackfillPanel;
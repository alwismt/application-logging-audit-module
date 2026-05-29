import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import Navbar from '../components/Navbar';
import { getAuditEvents, getLogs } from '../services/api';

export default function DashboardPage() {
  const [recentErrors, setRecentErrors] = useState([]);
  const [recentAudit, setRecentAudit] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      try {
        const [logsRes, auditRes] = await Promise.all([
          getLogs({ level: 'ERROR', page: 1, limit: 5 }),
          getAuditEvents({ page: 1, limit: 5 }),
        ]);
        setRecentErrors(logsRes.data || []);
        setRecentAudit(auditRes.data || []);
      } catch (err) {
        setError(err.message || 'Failed to load dashboard');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div className="page">
      <Navbar />
      <main className="content">
        <h1>Dashboard</h1>
        <p>Welcome to the Logging &amp; Audit admin console.</p>

        <div className="card-grid">
          <Link to="/logs" className="card">
            <h2>Logs</h2>
            <p>View and filter application logs</p>
          </Link>
          <Link to="/audit" className="card">
            <h2>Audit Events</h2>
            <p>Review security and activity audit trail</p>
          </Link>
        </div>

        {loading && <p className="loading">Loading...</p>}
        {error && <div className="error-banner">{error}</div>}

        {!loading && !error && (
          <>
            <section>
              <h2>Recent Errors</h2>
              {recentErrors.length === 0 ? (
                <p className="empty-state">No recent error logs.</p>
              ) : (
                <ul className="simple-list">
                  {recentErrors.map((log) => (
                    <li key={log.id}>
                      <strong>{log.level}</strong> — {log.message}
                    </li>
                  ))}
                </ul>
              )}
            </section>
            <section>
              <h2>Recent Audit Events</h2>
              {recentAudit.length === 0 ? (
                <p className="empty-state">No recent audit events.</p>
              ) : (
                <ul className="simple-list">
                  {recentAudit.map((event) => (
                    <li key={event.id}>
                      <strong>{event.action}</strong> — {event.username || 'unknown'} (
                      {event.status})
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </>
        )}
      </main>
    </div>
  );
}

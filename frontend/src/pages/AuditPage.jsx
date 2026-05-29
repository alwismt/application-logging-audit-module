import { useCallback, useEffect, useState } from 'react';
import Navbar from '../components/Navbar';
import AuditTable from '../components/AuditTable';
import { exportAuditEvents, getAuditEvents } from '../services/api';

const emptyFilters = {
  username: '',
  action: '',
  status: '',
  resource_type: '',
  from: '',
  to: '',
  page: 1,
  limit: 20,
};

export default function AuditPage() {
  const [filters, setFilters] = useState(emptyFilters);
  const [applied, setApplied] = useState(emptyFilters);
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const loadEvents = useCallback(async (params) => {
    setLoading(true);
    setError('');
    try {
      const query = Object.fromEntries(
        Object.entries(params).filter(([, v]) => v !== '' && v != null),
      );
      const result = await getAuditEvents(query);
      setEvents(result.data || []);
    } catch (err) {
      setError(err.message || 'Failed to load audit events');
      setEvents([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadEvents(applied);
  }, [applied, loadEvents]);

  function applyFilters(event) {
    event.preventDefault();
    setApplied({ ...filters, page: 1 });
  }

  function changePage(delta) {
    setApplied((prev) => ({ ...prev, page: Math.max(1, prev.page + delta) }));
    setFilters((prev) => ({ ...prev, page: Math.max(1, prev.page + delta) }));
  }

  async function handleExport(format) {
    const params = Object.fromEntries(
      Object.entries(applied).filter(([k, v]) => v !== '' && k !== 'page' && k !== 'limit'),
    );
    params.format = format;
    try {
      await exportAuditEvents(params);
    } catch (err) {
      setError(err.message || 'Export failed');
    }
  }

  return (
    <div className="page">
      <Navbar />
      <main className="content">
        <h1>Audit Events</h1>

        <form className="filters" onSubmit={applyFilters}>
          <label>
            Username
            <input
              type="text"
              value={filters.username}
              onChange={(e) => setFilters({ ...filters, username: e.target.value })}
            />
          </label>
          <label>
            Action
            <input
              type="text"
              value={filters.action}
              onChange={(e) => setFilters({ ...filters, action: e.target.value })}
              placeholder="LOGIN"
            />
          </label>
          <label>
            Status
            <select
              value={filters.status}
              onChange={(e) => setFilters({ ...filters, status: e.target.value })}
            >
              <option value="">All</option>
              <option value="SUCCESS">SUCCESS</option>
              <option value="FAILURE">FAILURE</option>
            </select>
          </label>
          <label>
            Resource Type
            <input
              type="text"
              value={filters.resource_type}
              onChange={(e) => setFilters({ ...filters, resource_type: e.target.value })}
            />
          </label>
          <label>
            From
            <input
              type="date"
              value={filters.from}
              onChange={(e) => setFilters({ ...filters, from: e.target.value })}
            />
          </label>
          <label>
            To
            <input
              type="date"
              value={filters.to}
              onChange={(e) => setFilters({ ...filters, to: e.target.value })}
            />
          </label>
          <button type="submit" className="btn-primary">
            Apply Filters
          </button>
          <button type="button" className="btn-secondary" onClick={() => handleExport('csv')}>
            Export CSV
          </button>
          <button type="button" className="btn-secondary" onClick={() => handleExport('json')}>
            Export JSON
          </button>
        </form>

        {loading && <p className="loading">Loading audit events...</p>}
        {error && <div className="error-banner">{error}</div>}
        {!loading && !error && <AuditTable events={events} />}

        <div className="pagination">
          <button type="button" className="btn-secondary" onClick={() => changePage(-1)}>
            Previous
          </button>
          <span>Page {applied.page}</span>
          <button
            type="button"
            className="btn-secondary"
            onClick={() => changePage(1)}
            disabled={events.length < applied.limit}
          >
            Next
          </button>
        </div>
      </main>
    </div>
  );
}

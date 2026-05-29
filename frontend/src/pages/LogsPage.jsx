import { useCallback, useEffect, useState } from 'react';
import Navbar from '../components/Navbar';
import LogTable from '../components/LogTable';
import { exportLogs, getLogs } from '../services/api';

const emptyFilters = {
  level: '',
  from: '',
  to: '',
  page: 1,
  limit: 20,
};

export default function LogsPage() {
  const [filters, setFilters] = useState(emptyFilters);
  const [applied, setApplied] = useState(emptyFilters);
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const loadLogs = useCallback(async (params) => {
    setLoading(true);
    setError('');
    try {
      const query = Object.fromEntries(
        Object.entries(params).filter(([, v]) => v !== '' && v != null),
      );
      const result = await getLogs(query);
      setLogs(result.data || []);
    } catch (err) {
      setError(err.message || 'Failed to load logs');
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadLogs(applied);
  }, [applied, loadLogs]);

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
      await exportLogs(params);
    } catch (err) {
      setError(err.message || 'Export failed');
    }
  }

  return (
    <div className="page">
      <Navbar />
      <main className="content">
        <h1>Application Logs</h1>

        <form className="filters" onSubmit={applyFilters}>
          <label>
            Level
            <select
              value={filters.level}
              onChange={(e) => setFilters({ ...filters, level: e.target.value })}
            >
              <option value="">All</option>
              <option value="INFO">INFO</option>
              <option value="WARNING">WARNING</option>
              <option value="ERROR">ERROR</option>
              <option value="DEBUG">DEBUG</option>
            </select>
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

        {loading && <p className="loading">Loading logs...</p>}
        {error && <div className="error-banner">{error}</div>}
        {!loading && !error && <LogTable logs={logs} />}

        <div className="pagination">
          <button type="button" className="btn-secondary" onClick={() => changePage(-1)}>
            Previous
          </button>
          <span>Page {applied.page}</span>
          <button
            type="button"
            className="btn-secondary"
            onClick={() => changePage(1)}
            disabled={logs.length < applied.limit}
          >
            Next
          </button>
        </div>
      </main>
    </div>
  );
}

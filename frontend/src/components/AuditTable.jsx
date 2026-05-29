export default function AuditTable({ events }) {
  if (!events.length) {
    return <p className="empty-state">No audit events found.</p>;
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Username</th>
            <th>Action</th>
            <th>Resource Type</th>
            <th>Status</th>
            <th>IP Address</th>
            <th>Timestamp</th>
          </tr>
        </thead>
        <tbody>
          {events.map((event) => (
            <tr key={event.id}>
              <td>{event.username || '-'}</td>
              <td>{event.action}</td>
              <td>{event.resource_type || '-'}</td>
              <td>
                <span className={`badge badge-${(event.status || '').toLowerCase()}`}>
                  {event.status}
                </span>
              </td>
              <td className="mono">{event.ip_address || '-'}</td>
              <td>{new Date(event.created_at).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

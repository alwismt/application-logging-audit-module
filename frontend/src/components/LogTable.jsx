export default function LogTable({ logs }) {
  if (!logs.length) {
    return <p className="empty-state">No logs found.</p>;
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Level</th>
            <th>Message</th>
            <th>Source</th>
            <th>Request ID</th>
            <th>Created At</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((log) => (
            <tr key={log.id}>
              <td>
                <span className={`badge badge-${(log.level || '').toLowerCase()}`}>
                  {log.level}
                </span>
              </td>
              <td>{log.message}</td>
              <td>{log.source || '-'}</td>
              <td className="mono">{log.request_id || '-'}</td>
              <td>{new Date(log.created_at).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

import { Link, useLocation } from 'react-router-dom';
import { clearToken } from '../services/api';

export default function Navbar() {
  const location = useLocation();

  function logout() {
    clearToken();
    window.location.href = '/login';
  }

  const links = [
    { to: '/dashboard', label: 'Dashboard' },
    { to: '/logs', label: 'Logs' },
    { to: '/audit', label: 'Audit Events' },
  ];

  return (
    <nav className="navbar">
      <div className="navbar-brand">Logging & Audit Admin</div>
      <div className="navbar-links">
        {links.map((link) => (
          <Link
            key={link.to}
            to={link.to}
            className={location.pathname === link.to ? 'active' : ''}
          >
            {link.label}
          </Link>
        ))}
      </div>
      <button type="button" className="btn-secondary" onClick={logout}>
        Logout
      </button>
    </nav>
  );
}

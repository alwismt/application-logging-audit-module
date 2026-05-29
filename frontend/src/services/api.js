const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const TOKEN_KEY = 'admin_token';

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

function redirectToLogin() {
  clearToken();
  if (window.location.pathname !== '/login') {
    window.location.href = '/login';
  }
}

export async function apiFetch(path, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };
  const token = getToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    redirectToLogin();
    throw new Error('Unauthorized');
  }

  return response;
}

export async function login(username, password) {
  const response = await fetch(`${API_URL}/admin/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || 'Login failed');
  }
  setToken(data.token);
  return data;
}

export async function getLogs(params = {}) {
  const query = new URLSearchParams(params).toString();
  const response = await apiFetch(`/admin/logs?${query}`);
  if (!response.ok) {
    throw new Error('Failed to load logs');
  }
  return response.json();
}

export async function getAuditEvents(params = {}) {
  const query = new URLSearchParams(params).toString();
  const response = await apiFetch(`/admin/audit-events?${query}`);
  if (!response.ok) {
    throw new Error('Failed to load audit events');
  }
  return response.json();
}

export async function exportResource(path, params = {}, filename = 'export') {
  const query = new URLSearchParams(params).toString();
  const response = await apiFetch(`${path}?${query}`);
  if (!response.ok) {
    throw new Error('Export failed');
  }
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
}

export function exportLogs(params) {
  const format = params.format || 'csv';
  return exportResource('/admin/logs/export', params, `logs.${format}`);
}

export function exportAuditEvents(params) {
  const format = params.format || 'csv';
  return exportResource('/admin/audit-events/export', params, `audit_events.${format}`);
}

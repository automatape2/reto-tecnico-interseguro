(() => {
  'use strict';

  const BASE_URL = window.APP_CONFIG.GO_API_BASE_URL;
  const TOKEN_KEY = 'interseguro_challenge_token';

  const el = (id) => document.getElementById(id);

  function getToken() {
    return localStorage.getItem(TOKEN_KEY) || '';
  }

  function setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
  }

  function setStatus(node, message, kind) {
    node.textContent = message;
    node.className = 'status' + (kind ? ' ' + kind : '');
  }

  function formatNumber(n) {
    if (n === null || n === undefined) return '—';
    return Number.isInteger(n) ? String(n) : n.toFixed(4);
  }

  function renderMatrixTable(matrix) {
    const table = document.createElement('table');
    table.className = 'matrix';
    const tbody = document.createElement('tbody');
    matrix.forEach((row) => {
      const tr = document.createElement('tr');
      row.forEach((value) => {
        const td = document.createElement('td');
        td.textContent = formatNumber(value);
        tr.appendChild(td);
      });
      tbody.appendChild(tr);
    });
    table.appendChild(tbody);
    return table;
  }

  function diagonalBadge(isDiagonal) {
    const span = document.createElement('span');
    span.className = 'badge ' + (isDiagonal ? 'yes' : 'no');
    span.textContent = isDiagonal ? 'Sí' : 'No';
    return span;
  }

  function renderPerMatrixStats(perMatrix) {
    const tbody = el('perMatrixTable').querySelector('tbody');
    tbody.innerHTML = '';
    Object.entries(perMatrix).forEach(([name, stats]) => {
      const tr = document.createElement('tr');
      const cells = [
        name,
        `${stats.rows} x ${stats.cols}`,
        formatNumber(stats.max),
        formatNumber(stats.min),
        formatNumber(stats.average),
        formatNumber(stats.sum),
      ];
      cells.forEach((text) => {
        const td = document.createElement('td');
        td.textContent = text;
        tr.appendChild(td);
      });
      const diagTd = document.createElement('td');
      diagTd.appendChild(diagonalBadge(stats.isDiagonal));
      tr.appendChild(diagTd);
      tbody.appendChild(tr);
    });
  }

  function renderOverallStats(overall) {
    const table = el('overallTable');
    table.innerHTML = '';
    const rows = [
      ['Cantidad de valores', overall.count],
      ['Máximo', formatNumber(overall.max)],
      ['Mínimo', formatNumber(overall.min)],
      ['Promedio', formatNumber(overall.average)],
      ['Suma total', formatNumber(overall.sum)],
      ['¿Alguna matriz es diagonal?', overall.anyDiagonal ? 'Sí' : 'No'],
      ['Matrices diagonales', overall.diagonalMatrices.length ? overall.diagonalMatrices.join(', ') : '(ninguna)'],
    ];
    rows.forEach(([label, value]) => {
      const tr = document.createElement('tr');
      const th = document.createElement('th');
      th.textContent = label;
      const td = document.createElement('td');
      td.textContent = value;
      tr.appendChild(th);
      tr.appendChild(td);
      table.appendChild(tr);
    });
  }

  function renderResults(body) {
    el('qTable').innerHTML = '';
    el('qTable').appendChild(renderMatrixTable(body.matrices.Q));
    el('rTable').innerHTML = '';
    el('rTable').appendChild(renderMatrixTable(body.matrices.R));
    renderPerMatrixStats(body.statistics.perMatrix);
    renderOverallStats(body.statistics.overall);
    el('results').style.display = 'block';
  }

  function parseMatrixInput(raw) {
    let parsed;
    try {
      parsed = JSON.parse(raw);
    } catch (e) {
      throw new Error('La matriz debe ser JSON válido, ej. [[1,2],[3,4]]');
    }
    if (!Array.isArray(parsed) || parsed.length === 0 || !Array.isArray(parsed[0])) {
      throw new Error('La matriz debe ser un array de arrays no vacío');
    }
    const cols = parsed[0].length;
    const rectangular = parsed.every((row) => Array.isArray(row) && row.length === cols);
    if (!rectangular) {
      throw new Error('Todas las filas deben tener la misma cantidad de columnas');
    }
    return parsed;
  }

  async function getTokenFromApiKey() {
    const apiKey = el('apiKey').value.trim();
    if (!apiKey) {
      setStatus(el('authStatus'), 'Ingresá el X-API-Key primero.', 'error');
      return;
    }
    setStatus(el('authStatus'), 'Solicitando token...', '');
    try {
      const resp = await fetch(`${BASE_URL}/api/v1/auth/token`, {
        method: 'POST',
        headers: { 'X-API-Key': apiKey },
      });
      const body = await resp.json();
      if (!resp.ok) {
        throw new Error(body.error ? body.error.message : `HTTP ${resp.status}`);
      }
      setToken(body.token);
      setStatus(el('authStatus'), 'Token obtenido y guardado.', 'ok');
    } catch (err) {
      setStatus(el('authStatus'), `Error: ${err.message}`, 'error');
    }
  }

  function useManualToken() {
    const token = el('manualToken').value.trim();
    if (!token) {
      setStatus(el('authStatus'), 'Pegá un token primero.', 'error');
      return;
    }
    setToken(token);
    setStatus(el('authStatus'), 'Token guardado.', 'ok');
  }

  async function submitMatrix() {
    const submitStatus = el('submitStatus');
    const token = getToken();
    if (!token) {
      setStatus(submitStatus, 'Primero obtené o pegá un token de autenticación.', 'error');
      return;
    }

    let matrix;
    try {
      matrix = parseMatrixInput(el('matrixInput').value);
    } catch (err) {
      setStatus(submitStatus, err.message, 'error');
      return;
    }

    setStatus(submitStatus, 'Calculando...', '');
    el('results').style.display = 'none';

    try {
      const resp = await fetch(`${BASE_URL}/api/v1/matrix/qr`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ matrix }),
      });
      const body = await resp.json();
      if (!resp.ok) {
        throw new Error(body.error ? `${body.error.code}: ${body.error.message}` : `HTTP ${resp.status}`);
      }
      renderResults(body);
      setStatus(submitStatus, `Listo (${body.meta.computedAtMs} ms).`, 'ok');
    } catch (err) {
      setStatus(submitStatus, `Error: ${err.message}`, 'error');
    }
  }

  el('getTokenBtn').addEventListener('click', getTokenFromApiKey);
  el('useManualTokenBtn').addEventListener('click', useManualToken);
  el('submitBtn').addEventListener('click', submitMatrix);
  el('loadExampleBtn').addEventListener('click', () => {
    el('matrixInput').value = '[[12, -51, 4], [6, 167, -68], [-4, 24, -41]]';
  });
})();

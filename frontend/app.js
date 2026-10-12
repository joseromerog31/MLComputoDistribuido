"use strict";

// Frontend de solo lectura para datos históricos en esta primera etapa.
// Todas las peticiones pasan por Nginx
// El navegador NO llama a workers directamente.
const API_BASE = "/api";
const HISTORY_KEY = "liga-mx-ml-history-v1";
const $ = (id) => document.getElementById(id);
const formatNumber = (v) => Number(v).toLocaleString("es-MX");
const formatMs = (v) => v < 1000 ? `${Math.round(v)} ms` : `${(v / 1000).toFixed(2)} s`;
let lastResponse = null;
let lastDurationMs = null;
let history = readHistory();
let isRunning = false;

function readHistory() {
  try {
    const saved = JSON.parse(localStorage.getItem(HISTORY_KEY) || "[]");
    return Array.isArray(saved) ? saved.slice(0, 15) : [];
  } catch { return []; }
}
function saveHistory() {
  try { localStorage.setItem(HISTORY_KEY, JSON.stringify(history)); }
  catch { /* localStorage puede estar deshabilitado. */ }
}
function node(tag, className, textValue) {
  const el = document.createElement(tag);
  if (className) el.className = className;
  if (textValue !== undefined) el.textContent = String(textValue);
  return el;
}
function showView(view) {
  for (const button of document.querySelectorAll("[data-view]")) {
    const active = button.dataset.view === view;
    button.classList.toggle("active", active);
    button.setAttribute("aria-current", active ? "page" : "false");
  }
  for (const section of document.querySelectorAll(".view")) section.classList.toggle("active", section.id === `view-${view}`);
  const labels = {dashboard:"Resumen",predictions:"Laboratorio ML",workers:"Workers",matches:"Partidos"};
  $("crumb-current").textContent = labels[view] || "Resumen";
}
function setNotice(message, type = "subtle") {
  const el = $("run-message");
  el.className = `notice ${type}`;
  el.textContent = message;
}
function dateTime(iso) {
  return new Date(iso).toLocaleString("es-MX", {day:"2-digit",month:"short",hour:"2-digit",minute:"2-digit"});
}

// AbortController impide que una solicitud de salud se quede colgada indefinidamente.
async function requestJSON(url, options = {}, timeoutMs = 180000) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  const start = performance.now();
  try {
    const response = await fetch(url, { ...options, signal: controller.signal, cache: "no-store" });
    const durationMs = performance.now() - start;
    const raw = await response.text();
    let data;
    try { data = raw ? JSON.parse(raw) : null; } catch { data = raw; }
    if (!response.ok) {
      const details = typeof data === "string" ? data : (data?.error || response.statusText);
      throw new Error(`HTTP ${response.status}: ${String(details).trim()}`);
    }
    return { data, durationMs, status: response.status };
  } finally { clearTimeout(timer); }
}

async function checkHealth() {
  const chip = $("top-health");
  try {
    const result = await requestJSON(`${API_BASE}/health`, {}, 8000);
    const healthy = result.data?.status === "ok";
    chip.className = `status-chip ${healthy ? "good" : "bad"}`;
    chip.innerHTML = `<span class="live-dot"></span> ${healthy ? "Coordinador conectado" : "Respuesta inesperada"}`;
    $("stat-health").textContent = healthy ? "En línea" : "Revisar";
    $("stat-health-details").textContent = `GET /api/health · ${formatMs(result.durationMs)}`;
  } catch (error) {
    chip.className = "status-chip bad";
    chip.innerHTML = `<span class="live-dot"></span> Sin conexión`;
    $("stat-health").textContent = "Sin conexión";
    $("stat-health-details").textContent = "Comprueba Docker y Nginx";
  }
}

function normalizeResponse(data) {
  // Coincide con coordinator/balancer/batch.go del repositorio original.
  if (!data || typeof data !== "object" || !Array.isArray(data.predictions)) {
    throw new Error("El servidor respondió, pero el JSON no tiene el formato esperado (predictions[]).");
  }
  return {
    total: Number.isFinite(Number(data.total_records)) ? Number(data.total_records) : data.predictions.length,
    workers: Number.isFinite(Number(data.workers_used)) ? Number(data.workers_used) : null,
    predictions: data.predictions,
    distribution: Array.isArray(data.distribution) ? data.distribution : [],
  };
}

async function runBatch(event) {
  if (event) event.preventDefault();
  if (isRunning) return;
  const input = $("batch-limit");
  const limit = Number(input.value);
  if (!Number.isInteger(limit) || limit < 1 || limit > 5000) {
    setNotice("El límite debe ser un número entero entre 1 y 5,000.", "error");
    input.focus();
    return;
  }
  isRunning = true;
  $("run-button").disabled = true;
  $("run-label").textContent = "Procesando lote…";
  setNotice(`Enviando POST /api/predict-batch con ${formatNumber(limit)} registros…`);
  const startedAt = new Date().toISOString();
  try {
    const result = await requestJSON(`${API_BASE}/predict-batch`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({limit}),
    });
    const parsed = normalizeResponse(result.data);
    lastResponse = result.data;
    lastDurationMs = result.durationMs;
    history.unshift({at:startedAt,requested:limit,processed:parsed.total,workers:parsed.workers,durationMs:result.durationMs,status:result.status});
    history = history.slice(0, 15);
    saveHistory();
    updateDashboard(parsed, result.durationMs);
    updatePredictionTable(parsed);
    updateWorkerCards(parsed);
    updateHistory();
    $("json-output").textContent = JSON.stringify(result.data, null, 2);
    setNotice(`HTTP ${result.status} · ${formatNumber(parsed.total)} registros procesados por ${parsed.workers ?? "—"} workers en ${formatMs(result.durationMs)}.`, "success");
  } catch (error) {
    setNotice(`No se pudo completar el lote: ${error.message}`, "error");
  } finally {
    isRunning = false;
    $("run-button").disabled = false;
    $("run-label").textContent = "▶  Ejecutar lote";
  }
}

function distributionRows(parsed) {
  return parsed.distribution.filter(x => x && typeof x.worker === "string" && Number.isFinite(Number(x.processed)));
}
function renderDistribution(target, parsed) {
  const rows = distributionRows(parsed);
  target.replaceChildren();
  if (!rows.length) {
    target.appendChild(node("p", "body-copy", "El backend no devolvió estadísticas por worker en distribution[]."));
    return;
  }
  const maximum = Math.max(...rows.map(x => Number(x.processed)), 1);
  for (const item of rows) {
    const wrapper = node("div", "dist-row");
    const meta = node("div", "dist-meta");
    meta.appendChild(node("span", "", item.worker));
    meta.appendChild(node("b", "", `${formatNumber(item.processed)} registros`));
    const track = node("div", "dist-bar-track");
    const fill = node("div", "dist-bar-fill");
    fill.style.width = `${Math.max(0, Number(item.processed)) / maximum * 100}%`;
    track.appendChild(fill);
    wrapper.append(meta, track);
    target.appendChild(wrapper);
  }
  target.appendChild(node("div", "dist-summary", `${formatNumber(parsed.total)} registros · ${parsed.workers ?? "—"} workers utilizados en esta ejecución`));
}
function updateDashboard(parsed, ms) {
  $("stat-records").textContent = formatNumber(parsed.total);
  $("stat-records-details").textContent = "Procesados en la última ejecución";
  $("stat-latency").textContent = formatMs(ms);
  $("lab-records").textContent = formatNumber(parsed.total);
  $("lab-workers").textContent = parsed.workers ?? "—";
  $("lab-latency").textContent = formatMs(ms);
  $("lab-predictions").textContent = formatNumber(parsed.predictions.length);
  $("distribution-tag").textContent = "Datos reales";
  renderDistribution($("distribution-chart"), parsed);
  renderDistribution($("worker-detail"), parsed);
  updateActivity();
}
function updatePredictionTable(parsed) {
  const tbody = $("prediction-rows");
  tbody.replaceChildren();
  for (const item of parsed.predictions.slice(0, 15)) {
    const tr = node("tr");
    const estimate = Number(item.prediction);
    tr.append(
      node("td", "", item.id ?? "—"),
      node("td", "prediction-number", Number.isFinite(estimate) ? estimate.toFixed(3) : "—"),
      node("td", "", "Regresión lineal"),
    );
    tbody.appendChild(tr);
  }
  if (parsed.predictions.length === 0) {
    const tr = node("tr"); const td = node("td","table-empty","El backend devolvió una lista de predicciones vacía."); td.colSpan = 3; tr.appendChild(td); tbody.appendChild(tr);
  }
  $("prediction-table-note").textContent = `Se muestran ${Math.min(parsed.predictions.length, 15)} de ${formatNumber(parsed.predictions.length)} predicciones. El JSON contiene la respuesta completa.`;
}
function updateWorkerCards(parsed) {
  const used = new Map(distributionRows(parsed).map(item => [item.worker, Number(item.processed)]));
  const cards = $("worker-cards");
  cards.replaceChildren();
  for (let i = 1; i <= 3; i++) {
    const name = `worker-${i}`;
    const card = node("article", "worker-card");
    card.appendChild(node("span", "worker-avatar", `0${i}`));
    card.appendChild(node("strong", "", name));
    let state = "No participó en el último lote";
    if (used.has(name)) state = `${formatNumber(used.get(name))} registros en el último lote`;
    card.appendChild(node("span", used.has(name) ? "worker-active" : "worker-inactive", state));
    cards.appendChild(card);
  }
}
function updateActivity() {
  const container = $("activity-list");
  container.replaceChildren();
  if (!history.length) { container.appendChild(node("div", "empty-inline", "No hay actividad registrada todavía.")); return; }
  for (const item of history.slice(0, 4)) {
    const row = node("div", "activity-row");
    row.appendChild(node("span", "activity-method", "POST"));
    const detail = node("div", "activity-content");
    detail.appendChild(node("strong", "", `/predict-batch · ${formatNumber(item.requested)} solicitados`));
    detail.appendChild(node("span", "", `${dateTime(item.at)} · ${formatMs(item.durationMs)}`));
    row.appendChild(detail);
    row.appendChild(node("span", "activity-status", `HTTP ${item.status}`));
    container.appendChild(row);
  }
}
function updateHistory() {
  const tbody = $("history-rows");
  tbody.replaceChildren();
  if (!history.length) {
    const tr = node("tr"); const td = node("td", "table-empty", "Todavía no tienes ejecuciones registradas."); td.colSpan = 5; tr.appendChild(td); tbody.appendChild(tr); return;
  }
  for (const item of history) {
    const tr = node("tr");
    for (const value of [dateTime(item.at),formatNumber(item.requested),formatNumber(item.processed),item.workers ?? "—",formatMs(item.durationMs)]) tr.appendChild(node("td", "", value));
    tbody.appendChild(tr);
  }
}

for (const button of document.querySelectorAll("[data-view]")) button.addEventListener("click", () => showView(button.dataset.view));
for (const button of document.querySelectorAll("[data-limit]")) {
  button.addEventListener("click", () => {
    $("batch-limit").value = button.dataset.limit;
    for (const preset of document.querySelectorAll("[data-limit]")) preset.classList.toggle("active", preset === button);
  });
}
$("batch-limit").addEventListener("input", () => {
  for (const preset of document.querySelectorAll("[data-limit]")) preset.classList.toggle("active", preset.dataset.limit === $("batch-limit").value);
});
$("batch-form").addEventListener("submit", runBatch);
$("refresh-health").addEventListener("click", checkHealth);
$("open-run").addEventListener("click", () => {showView("dashboard");$("batch-limit").focus();});
$("copy-json").addEventListener("click", async () => {
  if (!lastResponse) return;
  try { await navigator.clipboard.writeText(JSON.stringify(lastResponse, null, 2)); $("copy-json").textContent = "Copiado"; }
  catch { $("copy-json").textContent = "No se pudo copiar"; }
});
$("clear-history").addEventListener("click", () => {
  if (!window.confirm("¿Limpiar el historial de pruebas guardado en este navegador?")) return;
  history = [];
  saveHistory();
  updateHistory();
  updateActivity();
});
function updateClock() { $("header-time").textContent = new Date().toLocaleTimeString("es-MX", {hour:"2-digit",minute:"2-digit"}); }
updateClock();
setInterval(updateClock, 60000);
updateHistory();
updateActivity();
checkHealth();
setInterval(checkHealth, 15000);

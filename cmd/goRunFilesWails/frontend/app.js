const api = window.go?.main?.GUI;

const textSizeDebug = document.getElementsByClassName('textSizeDebug')?.[0];
const html = document.getElementsByTagName('html')?.[0];
let   fontSizeHtml = '10px';
const elUpdated                 = document.getElementById("updated");
const elVersion                 = document.getElementById("version");
const elNetStatus               = document.getElementById("netStatus");
const elNetDebug                = document.getElementById("netDebug");
const processCards              = document.getElementById("processCards");
const reloadBtn                 = document.getElementById("reloadConfig");
const saveBtn                   = document.getElementById("saveConfig");
const toggleBtn                 = document.getElementById("toggleConfig");
const lockConfigBtn             = document.getElementById("lockConfigBtn");
const restartAllBtn             = document.getElementById("restartAll");
const restartAutoManualBtn      = document.getElementById("restartAutoManual");
const stopAllBtn                = document.getElementById("stopAll");
const killCMDBtn                = document.getElementById("killCMD");
const toggleCheckProcessBtn     = document.getElementById("toggleCheckProcess");
const killNodeBtn               = document.getElementById("killNode");
const toggleConsoleBtn          = document.getElementById("toggleConsole");
const addProcessBtn             = document.getElementById("addProcess");
const configPanel               = document.getElementById("configPanel");
const configPassword            = document.getElementById("configPassword");
const unlockBtn                 = document.getElementById("unlockConfig");
const authModal                 = document.getElementById("authModal");
const closeAuth                 = document.getElementById("closeAuth");
const cancelAuth                = document.getElementById("cancelAuth");
const configModal               = document.getElementById("configModal");
const closeConfig               = document.getElementById("closeConfig");
const lockConfigModalBtn        = document.getElementById("lockConfigModalBtn");
const processDrawer             = document.getElementById("processDrawer");
const closeDrawer               = document.getElementById("closeDrawer");
const drawerTitle               = document.getElementById("drawerTitle");
const drawerStatusChip          = document.getElementById("drawerStatusChip");
const drawerStatusText          = document.getElementById("drawerStatusText");
const drawerOpenFolder          = document.getElementById("drawerOpenFolder");
const drawerRestart             = document.getElementById("drawerRestart");
const drawerStop                = document.getElementById("drawerStop");
const drawerStart               = document.getElementById("drawerStart");
const drawerPid                 = document.getElementById("drawerPid");
const drawerStarted             = document.getElementById("drawerStarted");
const drawerUptime              = document.getElementById("drawerUptime");
const drawerTarget              = document.getElementById("drawerTarget");
const drawerCpu                 = document.getElementById("drawerCpu");
const drawerGpu                 = document.getElementById("drawerGpu");
const drawerMem                 = document.getElementById("drawerMem");
const drawerNet                 = document.getElementById("drawerNet");
const drawerIo                  = document.getElementById("drawerIo");
const drawerName                = document.getElementById("drawerName");
const drawerDisabled            = document.getElementById("drawerDisabled");
const drawerType                = document.getElementById("drawerType");
const drawerProcess             = document.getElementById("drawerProcess");
const drawerPath                = document.getElementById("drawerPath");
const drawerCommand             = document.getElementById("drawerCommand");
const drawerArgs                = document.getElementById("drawerArgs");
const drawerScreen              = document.getElementById("drawerScreen");
const drawerScreenPicker        = document.getElementById("drawerScreenPicker");
const drawerCheckProcess        = document.getElementById("drawerCheckProcess");
const drawerCheckCmdline        = document.getElementById("drawerCheckCmdline");
const drawerCheckCmdlineExclude = document.getElementById("drawerCheckCmdlineExclude");
const drawerDelayStartTime      = document.getElementById("drawerDelayStartTime");
const drawerMonitorHang         = document.getElementById("drawerMonitorHang");
const drawerHangTimeout         = document.getElementById("drawerHangTimeout");
const drawerSave                = document.getElementById("drawerSave");
const errorConsoleContainer     = document.getElementById("errorConsoleContainer");
const errorConsole              = document.getElementById("errorConsole");
const schedulerTask             = document.getElementById("schedulerTask");
const schedulerState            = document.getElementById("schedulerState");
const schedulerLastRun          = document.getElementById("schedulerLastRun");
const schedulerLastResult       = document.getElementById("schedulerLastResult");
const schedulerNote             = document.getElementById("schedulerNote");
const installSchedulerBtn       = document.getElementById("installScheduler");
const removeSchedulerBtn        = document.getElementById("removeScheduler");
const refreshSchedulerBtn       = document.getElementById("refreshScheduler");

const cfgCheckTiming            = document.getElementById("cfgCheckTiming");
const cfgRestartTiming          = document.getElementById("cfgRestartTiming");
const cfgAutoRestart            = document.getElementById("cfgAutoRestart");
const cfgAutoRestartTime        = document.getElementById("cfgAutoRestartTime");
const cfgAutoRestartOnExit      = document.getElementById("cfgAutoRestartOnExit");
const cfgUseETWNetwork          = document.getElementById("cfgUseETWNetwork");
const cfgNetDebug               = document.getElementById("cfgNetDebug");
const cfgNetUnit                = document.getElementById("cfgNetUnit");
const cfgNetScale               = document.getElementById("cfgNetScale");
const cfgLaunchInNewConsole     = document.getElementById("cfgLaunchInNewConsole");
const cfgAutoCloseErrorDialogs  = document.getElementById("cfgAutoCloseErrorDialogs");
const cfgErrorWindowTitles      = document.getElementById("cfgErrorWindowTitles");
const cfgFind                   = document.getElementById("cfgFind");
const cfgProcesses              = document.getElementById("configProcesses");
const cfgScreens                = document.getElementById("cfgScreens");

let lastSnapshot = null;
const HISTORY_LEN = 40;
const metricHistory = new Map();
let sparkSeq = 0;
const ANIM_DURATION_MS = 320;
const MIN_TICK_MS = 100;
const ERROR_LOG_MAX = 600;
const errorLogLines = [];
const lastErrorByProcess = new Map();
let consoleOpened = false;
const CMD_CHECK_CMDLINE_EXCLUDE_DEFAULT = "jetbrains,js-language-service,typingsinstaller,eslint";
let availableScreens = [];
let tickIntervalMs = 500;
let tickTimer = null;
let tickInFlight = false;
let fontAnimTimer = null;
let checkProcessRunning = true;
const rowMap = new Map();

let currentConfigModel = null;
let pendingProcessOpen = "";
let pendingOpenConfig = false;
let activeProcessName = "";
let selectedProcessName = "";

const STATUS_META = {
  running: { label: "Running", icon: "▶", cls: "status-running" },
  started: { label: "Starting", icon: "◔", cls: "status-started" },
  stopped: { label: "Stopped", icon: "■", cls: "status-stopped" },
  disabled: { label: "Disabled", icon: "⏸", cls: "status-disabled" },
  hung: { label: "Hung", icon: "⚠", cls: "status-hung" },
  unknown: { label: "Unknown", icon: "?", cls: "status-unknown" },
};

const ICONS = {
  folder: `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6.5A2.5 2.5 0 0 1 5.5 4h4.2a2 2 0 0 1 1.5.7l1 1.3H18.5A2.5 2.5 0 0 1 21 8.5v7A2.5 2.5 0 0 1 18.5 18h-13A2.5 2.5 0 0 1 3 15.5v-9Z"/></svg>`,
  restart: `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5a7 7 0 1 1-6.1 10.4M7 5H3v4"/><path d="M3 9c1.4-3.4 4.7-5.8 8.5-5.8A8.5 8.5 0 1 1 5.2 18.5"/></svg>`,
  stop: `<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="6.5" y="6.5" width="11" height="11" rx="2.2"/></svg>`,
  start: `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5.5v13l11-6.5-11-6.5Z"/></svg>`,
  lock: `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M7.5 11V8.8a4.5 4.5 0 0 1 9 0V11"/><rect x="5" y="11" width="14" height="9" rx="2.2"/></svg>`,
};

const clamp = (v, min, max) => Math.max(min, Math.min(max, v));

const pushMetric = (name, cpu, gpu, mem, net, io) => {
  if (!metricHistory.has(name)) {
    metricHistory.set(name, { cpu: [], gpu: [], mem: [], net: [], io: [] });
  }
  const h = metricHistory.get(name);
  h.cpu.push(cpu);
  h.gpu.push(gpu);
  h.mem.push(mem);
  h.net.push(net);
  h.io.push(io);
  if (h.cpu.length > HISTORY_LEN) h.cpu.shift();
  if (h.gpu.length > HISTORY_LEN) h.gpu.shift();
  if (h.mem.length > HISTORY_LEN) h.mem.shift();
  if (h.net.length > HISTORY_LEN) h.net.shift();
  if (h.io.length > HISTORY_LEN) h.io.shift();
};

const buildSparkline = (values, color) => {
  const id = `grad-${sparkSeq++}`;
  const w = 90;
  const h = 26;
  const maxPoints = Math.max(values.length, 2);
  const step = w / (maxPoints - 1);
  const pts = values.map((v, i) => {
    const x = i * step;
    const y = h - (clamp(v, 0, 100) / 100) * h;
    return `${x.toFixed(2)},${y.toFixed(2)}`;
  });
  const poly = pts.join(" ");
  const area = `0,${h} ${poly} ${w},${h}`;
  return `
    <svg viewBox="0 0 ${w} ${h}" width="${w}" height="${h}" class="spark">
      <defs>
        <linearGradient id="${id}" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="${color}" stop-opacity="0.55" />
          <stop offset="100%" stop-color="${color}" stop-opacity="0" />
        </linearGradient>
      </defs>
      <polygon points="${area}" fill="url(#${id})" />
      <polyline points="${poly}" fill="none" stroke="${color}" stroke-width="2" />
    </svg>
  `;
};

const toFiniteOr = (v, fallback) => {
  const n = Number(v);
  return Number.isFinite(n) ? n : fallback;
};

const statusInfo = (it) => {
  if (!it) return STATUS_META.unknown;
  if (it.hung) return STATUS_META.hung;
  if (it.disabled || it.status === "disabled") return STATUS_META.disabled;
  return STATUS_META[it.status] || STATUS_META.unknown;
};

const escapeAttr = (value) => String(value ?? "")
  .replaceAll("&", "&amp;")
  .replaceAll('"', "&quot;")
  .replaceAll("<", "&lt;")
  .replaceAll(">", "&gt;");

const screenLabel = (screen) => {
  const parts = [
    `${screen.index}: ${screen.name || `Screen ${screen.index}`}`,
    `${screen.width}x${screen.height}`,
    `x=${screen.x}`,
    `y=${screen.y}`,
  ];
  if (screen.primary) parts.push("primary");
  return parts.join(" | ");
};

const renderScreenViz = (screens) => {
  if (!cfgScreens) return;

  if (!screens.length) {
    cfgScreens.textContent = "No screens detected";
    cfgScreens.style.height = "";
    return;
  }

  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  for (const s of screens) {
    if (s.x < minX) minX = s.x;
    if (s.y < minY) minY = s.y;
    if (s.x + s.width > maxX) maxX = s.x + s.width;
    if (s.y + s.height > maxY) maxY = s.y + s.height;
  }

  const virtW = maxX - minX;
  const virtH = maxY - minY;
  if (virtW <= 0 || virtH <= 0) {
    cfgScreens.textContent = screens.map((s) => screenLabel(s)).join("\n");
    cfgScreens.style.height = "";
    return;
  }

  const availW = Math.max(cfgScreens.clientWidth - 32, 200);
  const maxH = 260;
  const scale = Math.min(availW / virtW, maxH / virtH);

  cfgScreens.innerHTML = "";
  cfgScreens.style.position = "relative";
  cfgScreens.style.height = Math.round(virtH * scale) + "px";

  for (const s of screens) {
    const left = (s.x - minX) * scale;
    const top = (s.y - minY) * scale;
    const w = Math.max(s.width * scale - 4, 50);
    const h = Math.max(s.height * scale - 4, 34);

    const el = document.createElement("div");
    el.className = "monitor-viz";
    el.style.cssText = `left:${left}px;top:${top}px;width:${w}px;height:${h}px`;
    el.title = s.name || `Screen ${s.index}`;

    const num = document.createElement("div");
    num.className = "monitor-num";
    num.textContent = s.index;

    const nameEl = document.createElement("div");
    nameEl.className = "monitor-name";
    nameEl.textContent = s.name || `DISPLAY${s.index}`;

    const info = document.createElement("div");
    info.className = "monitor-info";
    info.textContent = `${s.width}x${s.height}`;

    const pos = document.createElement("div");
    pos.className = "monitor-pos";
    pos.textContent = `x=${s.x} y=${s.y}`;

    el.appendChild(num);
    el.appendChild(nameEl);
    el.appendChild(info);
    el.appendChild(pos);

    if (s.primary) {
      const badge = document.createElement("div");
      badge.className = "monitor-badge";
      badge.textContent = "PRIMARY";
      el.appendChild(badge);
    }

    cfgScreens.appendChild(el);
  }
};

const refreshScreens = async () => {
  if (!api?.GetScreens) {
    availableScreens = [];
    if (cfgScreens) cfgScreens.textContent = "Screen list is unavailable";
    return;
  }
  try {
    const screens = await api.GetScreens();
    availableScreens = Array.isArray(screens) ? screens : [];
    renderScreenViz(availableScreens);
  } catch (err) {
    availableScreens = [];
    if (cfgScreens) cfgScreens.textContent = err.message || String(err);
  }
};

const renderMonitorPicker = (container, selected) => {
  if (!container) return;

  if (!availableScreens.length) {
    container.textContent = "No screens";
    container.style.height = "";
    return;
  }

  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  for (const s of availableScreens) {
    if (s.x < minX) minX = s.x;
    if (s.y < minY) minY = s.y;
    if (s.x + s.width > maxX) maxX = s.x + s.width;
    if (s.y + s.height > maxY) maxY = s.y + s.height;
  }

  const virtW = maxX - minX;
  const virtH = maxY - minY;
  if (virtW <= 0 || virtH <= 0) {
    container.textContent = "No screens";
    container.style.height = "";
    return;
  }

  const availW = Math.max(Math.min(container.clientWidth || 240, 240), 80);
  const maxH = 70;
  const scale = Math.min(availW / virtW, maxH / virtH);

  container.innerHTML = "";
  container.style.position = "relative";
  container.style.height = Math.round(virtH * scale) + "px";

  let current = Number(selected) || 0;
  if (current <= 0) {
    const primary = availableScreens.find((s) => s.primary);
    if (primary) current = primary.index;
  }

  for (const s of availableScreens) {
    const left = (s.x - minX) * scale;
    const top = (s.y - minY) * scale;
    const w = Math.max(s.width * scale - 4, 24);
    const h = Math.max(s.height * scale - 4, 16);

    const el = document.createElement("div");
    el.className = "monitor-pick";
    el.style.cssText = `left:${left}px;top:${top}px;width:${w}px;height:${h}px`;
    el.title = s.name || `Screen ${s.index}`;

    if (current === s.index) {
      el.classList.add("selected");
    }

    const num = document.createElement("div");
    num.className = "monitor-pick-num";
    num.textContent = s.index;
    el.appendChild(num);

    if (s.primary) {
      const badge = document.createElement("div");
      badge.className = "monitor-pick-badge";
      badge.textContent = "P";
      el.appendChild(badge);
    }

    el.addEventListener("click", (e) => {
      e.stopPropagation();
      container.querySelectorAll(".monitor-pick").forEach((m) => m.classList.remove("selected"));
      el.classList.add("selected");
      const input = container.closest("label")?.querySelector('[data-f="screen"]');
      if (input) input.value = String(s.index);
    });

    container.appendChild(el);
  }

  const input = container.closest("label")?.querySelector('[data-f="screen"]');
  if (input) input.value = String(current);
};

const buildScreenOptions = (selected) => {
  const current = Number(selected) || 0;
  const opts = ['<option value="0">Default</option>'];
  for (const screen of availableScreens) {
    const index = Number(screen.index) || 0;
    opts.push(`<option value="${index}" ${current === index ? "selected" : ""}>${escapeAttr(screenLabel(screen))}</option>`);
  }
  if (current > 0 && !availableScreens.some((screen) => Number(screen.index) === current)) {
    opts.push(`<option value="${current}" selected>Screen ${current} (not detected)</option>`);
  }
  return opts.join("");
};

const animateNumber = (el, from, to, format, duration = ANIM_DURATION_MS) => {
  if (!el) return;
  const start = performance.now();
  const fromVal = toFiniteOr(from, toFiniteOr(to, 0));
  const toVal = toFiniteOr(to, fromVal);
  const delta = toVal - fromVal;

  if (!Number.isFinite(fromVal) || !Number.isFinite(toVal) || Math.abs(delta) < 0.0001) {
    el.textContent = format(toVal);
    return;
  }

  const easeOutCubic = (t) => 1 - Math.pow(1 - t, 3);
  const step = (now) => {
    const p = Math.min(1, (now - start) / duration);
    const eased = easeOutCubic(p);
    const value = fromVal + delta * eased;
    el.textContent = format(value);
    if (p < 1) {
      requestAnimationFrame(step);
    }
  };
  requestAnimationFrame(step);
};

const triggerFontSizeAnim = () => {
  if (!html) return;
  html.classList.remove("font-size-anim");
  void html.offsetWidth;
  html.classList.add("font-size-anim");
  if (fontAnimTimer) {
    clearTimeout(fontAnimTimer);
  }
  fontAnimTimer = setTimeout(() => {
    html.classList.remove("font-size-anim");
  }, 260);
};

const appendErrorLog = (line) => {
  if (!line) return;
  errorLogLines.unshift(line);
  if (errorLogLines.length > ERROR_LOG_MAX) {
    errorLogLines.splice(ERROR_LOG_MAX);
  }
  if (errorConsole) {
    errorConsole.value = errorLogLines.join("\n");
    if (consoleOpened) {
      errorConsole.scrollTop = 0;
    }
  }
};

const collectErrorLog = (data) => {
  const stamp = data?.updated || new Date().toISOString().replace("T", " ").slice(0, 19);
  const activeNames = new Set();
  for (const it of data.items || []) {
    const name = it.name || "unknown";
    activeNames.add(name);
    const currentError = (it.error || "").trim();
    const prevError = lastErrorByProcess.get(name) || "";
    if (currentError && currentError !== prevError) {
      appendErrorLog(`[${stamp}] ${name}: ${currentError}`);
    }
    lastErrorByProcess.set(name, currentError);
  }
  for (const name of Array.from(lastErrorByProcess.keys())) {
    if (!activeNames.has(name)) {
      lastErrorByProcess.delete(name);
    }
  }
};

const createCard = (name) => {
  const card = document.createElement("article");
  card.className = "process-card";
  card.dataset.name = name;
  card.innerHTML = `
    <div class="process-card__bg">
      <div class="metric-glow metric-glow--cpu"></div>
      <div class="metric-glow metric-glow--gpu"></div>
      <div class="metric-glow metric-glow--ram"></div>
      <div class="metric-glow metric-glow--net"></div>
      <div class="metric-glow metric-glow--io"></div>
    </div>
    <div class="process-card__top">
      <div class="process-card__identity">
        <div class="process-name"></div>
        <div class="process-meta">
          <span class="process-type"></span>
          <span class="process-status"></span>
        </div>
      </div>
      <div class="process-actions"></div>
    </div>
    <div class="process-card__main">
      <div class="process-stat"><span>PID</span><strong class="process-pid"></strong></div>
      <div class="process-stat"><span>Started</span><strong class="process-started"></strong></div>
      <div class="process-stat"><span>Uptime</span><strong class="process-uptime"></strong></div>
      <div class="process-stat full"><span>Target</span><strong class="process-target"></strong></div>
    </div>
    <div class="process-metrics">
      <div class="metric-chip"><span>CPU</span><strong class="metric-val anim-cpu"></strong><div class="spark-wrap"></div></div>
      <div class="metric-chip"><span>GPU</span><strong class="metric-val anim-gpu"></strong><div class="spark-wrap"></div></div>
      <div class="metric-chip"><span>RAM</span><strong class="metric-val anim-ram"></strong><div class="spark-wrap"></div></div>
      <div class="metric-chip"><span>NET</span><strong class="metric-val anim-net"></strong><div class="spark-wrap"></div></div>
      <div class="metric-chip"><span>IO</span><strong class="metric-val anim-io"></strong><div class="spark-wrap"></div></div>
    </div>
  `;

  const checkbox = document.createElement("input");
  checkbox.className = "action-switch";
  checkbox.type = "checkbox";
  checkbox.dataset.action = "toggle-disabled";
  checkbox.dataset.name = name;

  const label = document.createElement("label");
  label.className = "action-toggle neon-toggle";
  label.title = "Disabled";
  label.appendChild(checkbox);

  const actionWrap = card.querySelector(".process-actions");
  const makeBtn = (action, title, cls) => {
    const btn = document.createElement("button");
    btn.className = `neon-btn neon-btn--icon ${cls}`.trim();
    btn.dataset.action = action;
    btn.dataset.name = name;
    btn.title = title;
    btn.setAttribute("aria-label", title);
    btn.innerHTML = ICONS[action === "open-folder" ? "folder" : action] || "";
    return btn;
  };
  actionWrap.appendChild(label);
  actionWrap.appendChild(makeBtn("open-folder", "Open folder", "neon-btn--folder"));
  actionWrap.appendChild(makeBtn("restart", "Restart", "neon-btn--restart"));
  actionWrap.appendChild(makeBtn("stop", "Stop", "neon-btn--stop"));
  actionWrap.appendChild(makeBtn("start", "Start", "neon-btn--start"));

  const glows = card.querySelectorAll(".metric-glow");
  const sparkWraps = card.querySelectorAll(".spark-wrap");

  return {
    card,
    checkbox,
    btnStart: actionWrap.querySelector('[data-action="start"]'),
    nameEl: card.querySelector(".process-name"),
    typeEl: card.querySelector(".process-type"),
    statusEl: card.querySelector(".process-status"),
    pidEl: card.querySelector(".process-pid"),
    startedEl: card.querySelector(".process-started"),
    uptimeEl: card.querySelector(".process-uptime"),
    targetEl: card.querySelector(".process-target"),
    cpu: { val: card.querySelector(".anim-cpu"), spark: sparkWraps[0], glow: glows[0] },
    gpu: { val: card.querySelector(".anim-gpu"), spark: sparkWraps[1], glow: glows[1] },
    mem: { val: card.querySelector(".anim-ram"), spark: sparkWraps[2], glow: glows[2] },
    net: { val: card.querySelector(".anim-net"), spark: sparkWraps[3], glow: glows[3] },
    io: { val: card.querySelector(".anim-io"), spark: sparkWraps[4], glow: glows[4] },
  };
};

const updateCard = (row, it, prev, netUnit, netIsMB) => {
  row.card.classList.toggle("hung", !!it.hung);
  row.card.classList.toggle("row-disabled", !!it.disabled || it.status === "disabled");
  row.card.classList.toggle("status-running-bg", !it.hung && !it.disabled && it.status === "running");
  row.card.classList.toggle("status-started-bg", !it.hung && !it.disabled && it.status === "started");
  row.card.classList.toggle("status-stopped-bg", !it.hung && !it.disabled && (it.status === "stopped" || it.status === "unknown"));
  row.card.classList.toggle("status-disabled-bg", !!it.disabled || it.status === "disabled");

  row.checkbox.checked = !!it.disabled;
  row.btnStart.disabled = !(it.status !== "running" && it.status !== "started");

  const si = statusInfo(it);
  row.nameEl.textContent = it.name || "";
  row.typeEl.textContent = (it.type || "").toUpperCase();
  row.statusEl.className = `process-status ${si.cls}`;
  row.statusEl.textContent = `${si.icon} ${si.label}`;

  const pidNum = Number(it.pid);
  const prevPid = Number(prev.pid);
  if (Number.isFinite(pidNum) && pidNum > 0) {
    animateNumber(row.pidEl, Number.isFinite(prevPid) && prevPid > 0 ? prevPid : pidNum, pidNum, (v) => `${Math.max(0, Math.round(v))}`);
  } else {
    row.pidEl.textContent = "-";
  }
  row.startedEl.textContent = it.started_at || "-";
  row.uptimeEl.textContent = it.uptime || "-";
  row.targetEl.textContent = it.target || "";

  const cpuVal = parseFloat(it.cpu || "0") || 0;
  const gpuVal = parseFloat(it.gpu || "0") || 0;
  const memVal = parseFloat(it.mem_mb || "0") || 0;
  const netVal = parseFloat(it.net_kbs || "0") || 0;
  const ioVal = parseFloat(it.io_kbs || "0") || 0;
  const netKBVal = netIsMB ? netVal * 1024 : netVal;
  const netMBVal = netIsMB ? netVal : netVal / 1024;
  const ioKBVal = netIsMB ? ioVal * 1024 : ioVal;
  const ioMBVal = netIsMB ? ioVal : ioVal / 1024;

  pushMetric(it.name, cpuVal, gpuVal, memVal, netVal, ioVal);
  const hist = metricHistory.get(it.name) || { cpu: [], gpu: [], mem: [], net: [], io: [] };
  row.cpu.spark.innerHTML = buildSparkline(hist.cpu, "#67e8f9");
  row.gpu.spark.innerHTML = buildSparkline(hist.gpu, "#fca5a5");
  row.mem.spark.innerHTML = buildSparkline(hist.mem, "#a7f3d0");
  row.net.spark.innerHTML = buildSparkline(hist.net, "#c4b5fd");
  row.io.spark.innerHTML = buildSparkline(hist.io, "#f9d46b");

  animateNumber(row.cpu.val, toFiniteOr(parseFloat(prev.cpu || "0"), cpuVal), cpuVal, (v) => `${Math.max(0, Math.round(v))}%`);
  animateNumber(row.gpu.val, toFiniteOr(parseFloat(prev.gpu || "0"), gpuVal), gpuVal, (v) => `${Math.max(0, Math.round(v))}%`);
  animateNumber(row.mem.val, toFiniteOr(parseFloat(prev.mem_mb || "0"), memVal), memVal, (v) => `${Math.max(0, v).toFixed(2)}MB`);
  animateNumber(row.net.val, toFiniteOr(parseFloat(prev.net_kbs || "0"), netVal), netVal, (v) => netIsMB ? `${Math.max(0, v).toFixed(2)}${netUnit}` : `${Math.max(0, Math.round(v))}${netUnit}`);
  animateNumber(row.io.val, toFiniteOr(parseFloat(prev.io_kbs || "0"), ioVal), ioVal, (v) => netIsMB ? `${Math.max(0, v).toFixed(2)}${netUnit}` : `${Math.max(0, Math.round(v))}${netUnit}`);

  row.cpu.glow.style.opacity = `${Math.min(1, cpuVal / 100)}`;
  row.gpu.glow.style.opacity = `${Math.min(1, gpuVal / 100)}`;
  row.mem.glow.style.opacity = `${Math.min(1, memVal / 100)}`;
  row.net.glow.style.opacity = `${Math.min(1, Math.abs(netVal) / 250)}`;
  row.io.glow.style.opacity = `${Math.min(1, Math.abs(ioVal) / 250)}`;

  row.cpu.spark.parentElement.title = `CPU: ${cpuVal.toFixed(1)}%`;
  row.gpu.spark.parentElement.title = `GPU: ${gpuVal.toFixed(1)}%`;
  row.mem.spark.parentElement.title = `RAM: ${memVal.toFixed(2)} MB`;
  row.net.spark.parentElement.title = `NET: ${netKBVal.toFixed(1)} KB/s | ${netMBVal.toFixed(2)} MB/s`;
  row.io.spark.parentElement.title = `IO: ${ioKBVal.toFixed(1)} KB/s | ${ioMBVal.toFixed(2)} MB/s`;
};

const render = (data) => {
  if (!data) return;
  setCheckProcessButton(data.check_process_running !== false);
  elUpdated.textContent = data.updated || "—";
  elVersion.textContent = data.version || "—";
  if (elNetStatus) {
    const mode = data.net_mode || "—";
    const err = data.net_err || "";
    elNetStatus.textContent = err ? `${mode} (${err})` : mode;
    elNetStatus.title = err || "";
  }
  if (elNetDebug) {
    elNetDebug.textContent = data.net_dbg || "—";
  }
  collectErrorLog(data);
  const netUnit = (data.net_unit || "KB").toUpperCase();
  const netIsMB = netUnit === "MB";
  const prevMap = new Map();
  if (lastSnapshot && Array.isArray(lastSnapshot.items)) {
    for (const it of lastSnapshot.items) {
      prevMap.set(it.name, it);
    }
  }

  const windowSize = html?.getBoundingClientRect();
  if (windowSize) {
    const needFontSize = `${(windowSize.width / 2050) * 11}px`;
    if (needFontSize !== fontSizeHtml) {
      fontSizeHtml = needFontSize;
      if (textSizeDebug) {
        textSizeDebug.innerHTML = `${windowSize.width}x${windowSize.height}`;
      }
      html.style["font-size"] = fontSizeHtml;
      triggerFontSizeAnim();
    }
  }

  const seen = new Set();
  for (const it of data.items || []) {
    const name = it.name || "";
    if (!name) continue;
    seen.add(name);
    let row = rowMap.get(name);
    if (!row) {
      row = createCard(name);
      rowMap.set(name, row);
    }
    const prev = prevMap.get(name) || {};
    updateCard(row, it, prev, netUnit, netIsMB);
    processCards.appendChild(row.card);
  }

  for (const [name, row] of rowMap.entries()) {
    if (!seen.has(name)) {
      row.card.remove();
      rowMap.delete(name);
    }
  }
  if (selectedProcessName) {
    syncDrawer(selectedProcessName);
  }
  lastSnapshot = data;
};

const setCheckProcessButton = (running) => {
  checkProcessRunning = !!running;
  if (!toggleCheckProcessBtn) return;
  toggleCheckProcessBtn.textContent = checkProcessRunning ? "Stop Check Process" : "Start Check Process";
  toggleCheckProcessBtn.title = checkProcessRunning ? "Stop process checks" : "Start process checks";
  toggleCheckProcessBtn.setAttribute("aria-label", toggleCheckProcessBtn.title);
  toggleCheckProcessBtn.classList.toggle("active", !checkProcessRunning);
};

const tick = async () => {
  if (!api) return;
  const data = await api.GetSnapshot();
  render(data);
  if (data && Number.isFinite(data.check_timing_ms) && data.check_timing_ms > 0) {
    tickIntervalMs = Math.max(MIN_TICK_MS, data.check_timing_ms);
  }
};

processCards.addEventListener("click", async (e) => {
  const btn = e.target.closest("button");
  if (btn) {
    if (!api) return;
    const name = btn.dataset.name;
    const action = btn.dataset.action;
    try {
      if (action === "open-folder") await api.OpenFolder(name);
      if (action === "start") await api.Start(name);
      if (action === "stop") await api.Stop(name);
      if (action === "restart") await api.Restart(name);
    } catch (err) {
      console.error(err);
    }
    return;
  }
  if (e.target.closest("input, select, textarea, label, a")) return;
  const card = e.target.closest(".process-card[data-name]");
  if (!card) return;
  const name = card.dataset.name;
  if (!name) return;
  await openProcessEditor(name);
});

const openProcessEditor = async (name) => {
  if (!api) return;
  pendingProcessOpen = name;
  if (!isConfigUnlocked()) {
    openAuthModal();
    return;
  }
  await refreshScreens();
  const model = await api.GetConfigModel();
  currentConfigModel = model;
  openDrawer(name, model);
};

const focusProcessCard = (name) => {
  activeProcessName = name || "";
  for (const card of cfgProcesses.querySelectorAll(".process-card")) {
    card.classList.toggle("selected", card.dataset.name === activeProcessName);
  }
  for (const card of cfgProcesses.querySelectorAll(".process-card")) {
    if (card.dataset.name === activeProcessName) {
      card.scrollIntoView({ behavior: "smooth", block: "start" });
      break;
    }
  }
};

const findProcessModel = (name, model = currentConfigModel) => {
  if (!model || !Array.isArray(model.processes)) return null;
  return model.processes.find((p) => p.name === name) || null;
};

const drawerSetMetric = (el, value, title, colorClass) => {
  if (!el) return;
  el.textContent = value;
  if (title) el.title = title;
  if (colorClass) {
    el.className = "";
    el.classList.add(colorClass);
  }
};

const renderDrawerSpark = (container, values, color) => {
  if (!container) return;
  container.innerHTML = buildSparkline(values.length ? values : [0], color);
};

const openDrawer = (name, model = currentConfigModel) => {
  const p = findProcessModel(name, model);
  if (!p) return;
  selectedProcessName = name;
  const si = statusInfo(lastSnapshot?.items?.find((it) => it.name === name) || p);
  processDrawer.classList.remove("hidden");
  processDrawer.setAttribute("aria-hidden", "false");
  drawerTitle.textContent = p.name || name;
  drawerStatusChip.textContent = `${si.icon} ${si.label}`;
  drawerStatusChip.className = `drawer-chip ${si.cls}`;
  drawerStatusText.textContent = p.type ? `${p.type.toUpperCase()} • ${p.path || p.process || ""}` : (p.path || p.process || "—");
  drawerName.value = p.name || "";
  drawerDisabled.checked = !!p.disabled;
  drawerType.value = p.type || "exe";
  drawerProcess.value = p.process || "";
  drawerPath.value = p.path || "";
  drawerCommand.value = p.command || "";
  drawerArgs.value = p.args || "";
  drawerScreen.value = String(Number(p.screen) || 0);
  drawerCheckProcess.value = p.checkProcess || "";
  drawerCheckCmdline.value = p.checkCmdline || "";
  drawerCheckCmdlineExclude.value = p.checkCmdlineExclude || "";
  drawerDelayStartTime.value = p.delayStartTime || "";
  drawerMonitorHang.checked = !!p.monitorHang;
  drawerHangTimeout.value = p.hangTimeout || "";
  drawerPid.textContent = (lastSnapshot?.items?.find((it) => it.name === name)?.pid) || "—";
  drawerStarted.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.started_at || "—";
  drawerUptime.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.uptime || "—";
  drawerTarget.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.target || "—";
  drawerCpu.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.cpu ? `${lastSnapshot.items.find((it) => it.name === name).cpu}%` : "—";
  drawerGpu.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.gpu ? `${lastSnapshot.items.find((it) => it.name === name).gpu}%` : "—";
  drawerMem.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.mem_mb ? `${lastSnapshot.items.find((it) => it.name === name).mem_mb}MB` : "—";
  drawerNet.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.net_kbs ? `${lastSnapshot.items.find((it) => it.name === name).net_kbs}KB/s` : "—";
  drawerIo.textContent = lastSnapshot?.items?.find((it) => it.name === name)?.io_kbs ? `${lastSnapshot.items.find((it) => it.name === name).io_kbs}KB/s` : "—";
  const hist = metricHistory.get(name) || { cpu: [], gpu: [], mem: [], net: [], io: [] };
  renderDrawerSpark(drawerCpuSpark, hist.cpu, "#67e8f9");
  renderDrawerSpark(drawerGpuSpark, hist.gpu, "#fca5a5");
  renderDrawerSpark(drawerMemSpark, hist.mem, "#a7f3d0");
  renderDrawerSpark(drawerNetSpark, hist.net, "#c4b5fd");
  renderDrawerSpark(drawerIoSpark, hist.io, "#f9d46b");
  renderMonitorPicker(drawerScreenPicker, p.screen);
  focusProcessCard(name);
};

const syncDrawer = (name) => {
  const p = findProcessModel(name, currentConfigModel);
  if (!p) return;
  const it = lastSnapshot?.items?.find((x) => x.name === name) || p;
  const si = statusInfo(it);
  drawerStatusChip.textContent = `${si.icon} ${si.label}`;
  drawerStatusChip.className = `drawer-chip ${si.cls}`;
  drawerStatusText.textContent = p.type ? `${p.type.toUpperCase()} • ${p.path || p.process || ""}` : (p.path || p.process || "—");
  drawerPid.textContent = it.pid || "—";
  drawerStarted.textContent = it.started_at || "—";
  drawerUptime.textContent = it.uptime || "—";
  drawerTarget.textContent = it.target || "—";
  drawerCpu.textContent = it.cpu ? `${it.cpu}%` : "—";
  drawerGpu.textContent = it.gpu ? `${it.gpu}%` : "—";
  drawerMem.textContent = it.mem_mb ? `${it.mem_mb}MB` : "—";
  drawerNet.textContent = it.net_kbs ? `${it.net_kbs}KB/s` : "—";
  drawerIo.textContent = it.io_kbs ? `${it.io_kbs}KB/s` : "—";
}

const closeDrawerPanel = () => {
  processDrawer.classList.add("hidden");
  processDrawer.setAttribute("aria-hidden", "true");
  selectedProcessName = "";
};

const collectDrawerProcess = () => ({
  name: drawerName.value.trim(),
  disabled: drawerDisabled.checked,
  type: drawerType.value,
  process: drawerProcess.value,
  path: drawerPath.value,
  command: drawerCommand.value,
  args: drawerArgs.value,
  screen: Number(drawerScreen.value || 0),
  checkProcess: drawerCheckProcess.value,
  checkCmdline: drawerCheckCmdline.value,
  checkCmdlineExclude: drawerCheckCmdlineExclude.value,
  delayStartTime: drawerDelayStartTime.value,
  monitorHang: drawerMonitorHang.checked,
  hangTimeout: drawerHangTimeout.value,
});

const saveDrawerProcess = async () => {
  if (!api || !currentConfigModel) return;
  const updated = collectDrawerProcess();
  if (!updated.name) throw new Error("Name is required");
  const next = structuredClone(currentConfigModel);
  next.processes = Array.isArray(next.processes) ? next.processes : [];
  const idx = next.processes.findIndex((p) => p.name === selectedProcessName);
  if (idx < 0) throw new Error("Process not found");
  next.processes[idx] = updated;
  await api.SaveConfigModel(next);
  currentConfigModel = next;
  selectedProcessName = updated.name;
  await refreshScreens();
  renderConfig(next);
  openDrawer(updated.name, next);
  await tick();
};

const isConfigUnlocked = () => unlocked || localStorage.getItem("goRunFilesUnlocked") === "1";

const setConfigUnlocked = (value) => {
  unlocked = !!value;
  localStorage.setItem("goRunFilesUnlocked", unlocked ? "1" : "0");
};

processCards.addEventListener("change", async (e) => {
  const el = e.target;
  if (!el || !api) return;
  if (el.dataset.action !== "toggle-disabled") return;
  const name = el.dataset.name;
  const disabled = !!el.checked;
  const prev = !disabled;
  el.disabled = true;
  try {
    const model = await api.GetConfigModel();
    if (!model || !Array.isArray(model.processes)) {
      throw new Error("Config load failed");
    }
    for (const p of model.processes) {
      if (p.name === name) {
        p.disabled = disabled;
        break;
      }
    }
    await api.SaveConfigModel(model);
    await tick();
  } catch (err) {
    console.error(err);
    el.checked = prev;
    alert(err.message || String(err));
  } finally {
    el.disabled = false;
  }
});

closeDrawer.addEventListener("click", closeDrawerPanel);
processDrawer.addEventListener("click", (e) => {
  if (e.target.classList.contains("drawer-backdrop")) {
    closeDrawerPanel();
  }
});

[
  [drawerOpenFolder, "folder", "Open folder"],
  [drawerRestart, "restart", "Restart"],
  [drawerStop, "stop", "Stop"],
  [drawerStart, "start", "Start"],
].forEach(([el, key, label]) => {
  el.classList.add("neon-btn--icon");
  el.innerHTML = ICONS[key] || "";
  el.title = label;
  el.setAttribute("aria-label", label);
});

drawerOpenFolder.addEventListener("click", async () => {
  if (!api || !selectedProcessName) return;
  await api.OpenFolder(selectedProcessName);
});
drawerRestart.addEventListener("click", async () => {
  if (!api || !selectedProcessName) return;
  await api.Restart(selectedProcessName);
});
drawerStop.addEventListener("click", async () => {
  if (!api || !selectedProcessName) return;
  await api.Stop(selectedProcessName);
});
drawerStart.addEventListener("click", async () => {
  if (!api || !selectedProcessName) return;
  await api.Start(selectedProcessName);
});
drawerSave.addEventListener("click", async () => {
  if (!api) return;
  try {
    await saveDrawerProcess();
  } catch (err) {
    alert(err.message || String(err));
  }
});

restartAllBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    await api.RestartAll();
  } catch (err) {
    console.error(err);
  }
});

restartAutoManualBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    await api.RestartAutoManual();
  } catch (err) {
    console.error(err);
  }
});

stopAllBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    await api.StopAll();
  } catch (err) {
    console.error(err);
  }
});

killCMDBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    await api.KillCMD();
  } catch (err) {
    console.error(err);
  }
});

toggleCheckProcessBtn.addEventListener("click", async () => {
  if (!api) return;
  toggleCheckProcessBtn.disabled = true;
  try {
    const running = checkProcessRunning
      ? await api.StopCheckProcess()
      : await api.StartCheckProcess();
    setCheckProcessButton(running);
    await tick();
  } catch (err) {
    console.error(err);
  } finally {
    toggleCheckProcessBtn.disabled = false;
  }
});

killNodeBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    await api.KillNode();
  } catch (err) {
    console.error(err);
  }
});

toggleConsoleBtn.addEventListener("click", () => {
  consoleOpened = !consoleOpened;
  errorConsoleContainer.classList.toggle("is-open", consoleOpened);
  toggleConsoleBtn.classList.toggle("active", consoleOpened);
  if (consoleOpened && errorConsole) {
    errorConsole.scrollTop = 0;
  }
});

const applyFilter = () => {
  const filter = cfgFind.value.trim().toLowerCase();
  for (const card of cfgProcesses.querySelectorAll(".process-card")) {
    const name = (card.querySelector('[data-f="name"]')?.value || "").trim().toLowerCase();
    if (!filter || name.includes(filter)) {
      card.classList.remove("hidden-by-filter");
    } else {
      card.classList.add("hidden-by-filter");
    }
  }
};

const renderConfig = (model) => {
  if (!model) return;
  currentConfigModel = model;

  const s = model.settings || {};
  cfgCheckTiming.value = s.checkTiming || "";
  cfgRestartTiming.value = s.restartTiming || "";
  cfgAutoRestart.checked = !!s.autoRestart;
  cfgAutoRestartTime.value = s.autoRestartTime || "";
  cfgAutoRestartOnExit.checked = !!s.autoRestartOnExit;
  cfgUseETWNetwork.checked = !!s.useETWNetwork;
  cfgNetDebug.checked = !!s.netDebug;
  cfgNetUnit.value = (s.netUnit || "KB").toUpperCase();
  cfgNetScale.value = String(s.netScale || "1");
  cfgLaunchInNewConsole.checked = !!s.launchInNewConsole;
  cfgAutoCloseErrorDialogs.checked = !!s.autoCloseErrorDialogs;
  cfgErrorWindowTitles.value = s.errorWindowTitles || "";

  cfgProcesses.innerHTML = "";

  for (const p of model.processes || []) {
    cfgProcesses.appendChild(buildProcessRow(p));
  }

  applyFilter();
  if (activeProcessName) focusProcessCard(activeProcessName);
};

const buildProcessRow = (p = {}) => {
  const initialType = p.type || "exe";
  const initialExclude = p.checkCmdlineExclude || (initialType === "cmd" ? CMD_CHECK_CMDLINE_EXCLUDE_DEFAULT : "");
  const card = document.createElement("div");
  card.className = "process-card";
  card.dataset.name = p.name || "";
  card.innerHTML = `
    <div class="process-grid">
      <label>Name
        <input data-f="name" value="${escapeAttr(p.name)}" />
      </label>
      <label>Disabled
        <input data-f="disabled" type="checkbox" ${p.disabled ? "checked" : ""} />
      </label>
      <label>Type
        <select data-f="type">
          <option value="exe">exe</option>
          <option value="cmd">cmd</option>
          <option value="bat">bat</option>
        </select>
      </label>
      <label>Process
        <input data-f="process" value="${escapeAttr(p.process)}" />
      </label>
      <label>Path
        <input data-f="path" value="${escapeAttr(p.path)}" />
      </label>
      <label>Command
        <input data-f="command" value="${escapeAttr(p.command)}" />
      </label>
      <label>Args
        <input data-f="args" value="${escapeAttr(p.args)}" />
      </label>
      <label>Screen
        <input data-f="screen" type="hidden" value="${Number(p.screen) || 0}" />
        <div class="monitor-picker"></div>
      </label>
      <label>CheckProcess
        <input data-f="checkProcess" value="${escapeAttr(p.checkProcess)}" />
      </label>
      <label>CheckCmdline
        <input data-f="checkCmdline" value="${escapeAttr(p.checkCmdline)}" />
      </label>
      <label>CheckCmdlineExclude
        <input data-f="checkCmdlineExclude" value="${escapeAttr(initialExclude)}" />
      </label>
      <label>DelayStartTime
        <input data-f="delayStartTime" value="${escapeAttr(p.delayStartTime)}" placeholder="30s" />
      </label>
      <label>MonitorHang
        <input data-f="monitorHang" type="checkbox" ${p.monitorHang ? "checked" : ""} />
      </label>
      <label>HangTimeout
        <input data-f="hangTimeout" value="${escapeAttr(p.hangTimeout)}" />
      </label>
    </div>
    <div class="process-actions">
      <button data-action="remove">Remove</button>
    </div>
  `;
  const picker = card.querySelector('.monitor-picker');
  if (picker) renderMonitorPicker(picker, p.screen);

  const typeSelect = card.querySelector('select[data-f="type"]');
  typeSelect.value = initialType;
  typeSelect.addEventListener("change", () => {
    const excludeInput = card.querySelector('input[data-f="checkCmdlineExclude"]');
    if (!excludeInput) return;
    if (typeSelect.value === "cmd" && !(excludeInput.value || "").trim()) {
      excludeInput.value = CMD_CHECK_CMDLINE_EXCLUDE_DEFAULT;
    }
  });
  return card;
};

const collectConfig = () => {
  const processes = [];
  const names = new Set();
  for (const card of cfgProcesses.querySelectorAll(".process-card")) {
    const get = (f) => card.querySelector(`[data-f="${f}"]`);
    const name = (get("name").value || "").trim();
    if (!name) continue;
    if (names.has(name)) throw new Error(`Duplicate process name: ${name}`);
    names.add(name);
    processes.push({
      name,
      disabled: get("disabled").checked,
      type: get("type").value,
      process: get("process").value,
      path: get("path").value,
      command: get("command").value,
      args: get("args").value,
      screen: Number(get("screen")?.value || 0),
      checkProcess: get("checkProcess").value,
      checkCmdline: get("checkCmdline").value,
      checkCmdlineExclude: get("checkCmdlineExclude").value,
      delayStartTime: get("delayStartTime").value,
      monitorHang: get("monitorHang").checked,
      hangTimeout: get("hangTimeout").value,
    });
  }
  return {
    settings: {
      checkTiming: cfgCheckTiming.value,
      restartTiming: cfgRestartTiming.value,
      autoRestart: cfgAutoRestart.checked,
      autoRestartTime: cfgAutoRestartTime.value,
      autoRestartOnExit: cfgAutoRestartOnExit.checked,
      useETWNetwork: cfgUseETWNetwork.checked,
      netDebug: cfgNetDebug.checked,
      netUnit: cfgNetUnit.value,
      netScale: cfgNetScale.value,
      launchInNewConsole: cfgLaunchInNewConsole.checked,
      autoCloseErrorDialogs: cfgAutoCloseErrorDialogs.checked,
      errorWindowTitles: cfgErrorWindowTitles.value,
      cfgFind: cfgFind.value,
    },
    processes,
  };
};

reloadBtn.addEventListener("click", async () => {
  if (!api) return;
  if (!unlocked) return;
  await refreshScreens();
  const model = await api.GetConfigModel();
  renderConfig(model);
});

cfgFind.addEventListener("input", () => {
  applyFilter();
});

cfgFind.addEventListener("keydown", (e) => {
  if (e.key === "Escape") {
    cfgFind.value = "";
    applyFilter();
  }
});

saveBtn.addEventListener("click", async () => {
  if (!api) return;
  if (!unlocked) return;
  try {
    const model = collectConfig();

    await api.SaveConfigModel(model);
  } catch (err) {
    alert(err.message || String(err));
  }
});

toggleBtn.addEventListener("click", () => {
  if (isConfigUnlocked()) {
    openConfigModal();
    return;
  }
  pendingOpenConfig = true;
  openAuthModal();
});

const scheduleTick = (delayMs) => {
  if (tickTimer) {
    clearTimeout(tickTimer);
  }
  tickTimer = setTimeout(runTick, delayMs);
};

const runTick = async () => {
  if (tickInFlight || !api) {
    scheduleTick(tickIntervalMs);
    return;
  }
  tickInFlight = true;
  const start = performance.now();
  try {
    await tick();
  } finally {
    tickInFlight = false;
    const elapsed = performance.now() - start;
    const delay = Math.max(0, tickIntervalMs - elapsed);
    scheduleTick(delay);
  }
};

window.onload = async () => {
  await tick();
  unlocked = isConfigUnlocked();
  if (api && unlocked) {
    await refreshScreens();
    const model = await api.GetConfigModel();
    renderConfig(model);
  }
  await refreshSchedulerStatus();
  if (!unlocked) {
    lockConfig();
  } else {
    closeAuthModal();
  }
  scheduleTick(tickIntervalMs);
};

addProcessBtn.addEventListener("click", () => {
  if (!unlocked) return;
  cfgProcesses.appendChild(buildProcessRow({}));
});

cfgProcesses.addEventListener("click", (e) => {
  const btn = e.target.closest("button");
  if (!btn) return;
  if (btn.dataset.action === "remove") {
    if (!unlocked) return;
    btn.closest(".process-card").remove();
  }
});

let unlocked = false;
const PASSWORD = "art3d";

const openConfigModal = async () => {
  if (!isConfigUnlocked()) {
    pendingOpenConfig = true;
    openAuthModal();
    return;
  }
  configModal.classList.remove("hidden");
  configPanel.classList.remove("hidden");
  document.querySelector(".config-actions").classList.remove("hidden");
  document.querySelector(".config-grid").classList.remove("hidden");
  document.querySelector(".process-list").classList.remove("hidden");
  await refreshScreens();
  const model = currentConfigModel || await api.GetConfigModel();
  renderConfig(model);
};

const lockConfig = () => {
  setConfigUnlocked(false);
  pendingOpenConfig = false;
  document.querySelector(".config-grid").classList.add("hidden");
  document.querySelector(".process-list").classList.add("hidden");
  document.querySelector(".config-actions").classList.add("hidden");
  configPassword.value = "";
  configPanel.classList.add("hidden");
  configModal.classList.add("hidden");
  closeDrawerPanel();
  activeProcessName = "";
};

const unlockConfig = async () => {
  if (configPassword.value !== PASSWORD) {
    alert("Неверный пароль");
    return;
  }
  setConfigUnlocked(true);
  closeAuthModal();
  if (pendingOpenConfig) {
    pendingOpenConfig = false;
    await openConfigModal();
  }
  if (pendingProcessOpen) {
    const model = currentConfigModel || await api.GetConfigModel();
    currentConfigModel = model;
    openDrawer(pendingProcessOpen, model);
    pendingProcessOpen = "";
  }
  await refreshSchedulerStatus();
};

unlockBtn.addEventListener("click", unlockConfig);

lockConfigBtn.addEventListener("click", lockConfig);
lockConfigModalBtn.addEventListener("click", lockConfig);

const openAuthModal = () => {
  authModal.classList.remove("hidden");
  configPassword.focus();
};

const closeAuthModal = () => {
  authModal.classList.add("hidden");
  configPassword.value = "";
};

closeAuth.addEventListener("click", closeAuthModal);
cancelAuth.addEventListener("click", closeAuthModal);
authModal.addEventListener("click", (e) => {
  if (e.target.classList.contains("modal-backdrop")) {
    closeAuthModal();
  }
});

closeConfig.addEventListener("click", () => {
  lockConfig();
});
configModal.addEventListener("click", (e) => {
  if (e.target.classList.contains("modal-backdrop")) {
    lockConfig();
  }
});

const renderSchedulerStatus = (s) => {
  if (!s || !schedulerTask || !schedulerState || !schedulerLastRun || !schedulerLastResult || !schedulerNote) return;
  schedulerTask.textContent = s.taskName || "—";
  if (!s.installed) {
    schedulerState.textContent = "Not installed";
  } else {
    schedulerState.textContent = s.running ? "Running" : (s.state || "Installed");
  }
  schedulerLastRun.textContent = s.lastRunTime || "—";
  schedulerLastResult.textContent = Number.isFinite(s.lastTaskResult) ? String(s.lastTaskResult) : "—";
  const noteParts = [];
  if (s.installed) {
    noteParts.push("Watchdog restarts app on exit.");
  } else {
    noteParts.push("Install to auto-start with admin rights.");
  }
  if (s.error) {
    noteParts.push(s.error);
  }
  schedulerNote.textContent = noteParts.join(" ");
};

const refreshSchedulerStatus = async () => {
  if (!api) return;
  try {
    const s = await api.GetSchedulerStatus();
    renderSchedulerStatus(s);
  } catch (err) {
    renderSchedulerStatus({ taskName: "—", error: err.message || String(err) });
  }
};

if (installSchedulerBtn) {
  installSchedulerBtn.addEventListener("click", async () => {
    if (!api) return;
    try {
      await api.InstallScheduler();
      await refreshSchedulerStatus();
    } catch (err) {
      alert(err.message || String(err));
    }
  });
}

if (removeSchedulerBtn) {
  removeSchedulerBtn.addEventListener("click", async () => {
    if (!api) return;
    try {
      await api.RemoveScheduler();
      await refreshSchedulerStatus();
    } catch (err) {
      alert(err.message || String(err));
    }
  });
}

if (refreshSchedulerBtn) {
  refreshSchedulerBtn.addEventListener("click", refreshSchedulerStatus);
}


// Checkbox pulse animation
document.addEventListener("change", (e) => {
  const el = e.target;
  if (el && el.type === "checkbox") {
    el.classList.remove("pulse");
    void el.offsetWidth;
    el.classList.add("pulse");
  }
});

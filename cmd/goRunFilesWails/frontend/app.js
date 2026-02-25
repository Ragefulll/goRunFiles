const api = window.go?.main?.GUI;

const elUpdated = document.getElementById("updated");
const elVersion = document.getElementById("version");
const elNetStatus = document.getElementById("netStatus");
const elNetDebug = document.getElementById("netDebug");
const tbody = document.getElementById("tbody");
const reloadBtn = document.getElementById("reloadConfig");
const saveBtn = document.getElementById("saveConfig");
const toggleBtn = document.getElementById("toggleConfig");
const restartAllBtn = document.getElementById("restartAll");
const killCMDBtn = document.getElementById("killCMD");
const killNodeBtn = document.getElementById("killNode");
const addProcessBtn = document.getElementById("addProcess");
const configPanel = document.getElementById("configPanel");
const configPassword = document.getElementById("configPassword");
const unlockBtn = document.getElementById("unlockConfig");
const authModal = document.getElementById("authModal");
const closeAuth = document.getElementById("closeAuth");
const cancelAuth = document.getElementById("cancelAuth");
const configModal = document.getElementById("configModal");
const closeConfig = document.getElementById("closeConfig");

const cfgCheckTiming = document.getElementById("cfgCheckTiming");
const cfgRestartTiming = document.getElementById("cfgRestartTiming");
const cfgUseETWNetwork = document.getElementById("cfgUseETWNetwork");
const cfgNetDebug = document.getElementById("cfgNetDebug");
const cfgNetUnit = document.getElementById("cfgNetUnit");
const cfgNetScale = document.getElementById("cfgNetScale");
const cfgLaunchInNewConsole = document.getElementById("cfgLaunchInNewConsole");
const cfgAutoCloseErrorDialogs = document.getElementById("cfgAutoCloseErrorDialogs");
const cfgErrorWindowTitles = document.getElementById("cfgErrorWindowTitles");
const cfgProcesses = document.getElementById("configProcesses");

let lastSnapshot = null;
const HISTORY_LEN = 40;
const metricHistory = new Map();
let sparkSeq = 0;
const ANIM_DURATION_MS = 320;

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

const render = (data) => {
  if (!data) return;
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
  const netUnit = (data.net_unit || "KB").toUpperCase();
  tbody.innerHTML = "";

  const prevMap = new Map();
  if (lastSnapshot && Array.isArray(lastSnapshot.items)) {
    for (const it of lastSnapshot.items) {
      prevMap.set(it.name, it);
    }
  }

  for (const it of data.items || []) {
    const prev = prevMap.get(it.name) || {};
    const tr = document.createElement("tr");
    if (it.hung) tr.classList.add("hung");
    if (it.status === "disabled") tr.classList.add("row-disabled");
    const canStart = it.status !== "running" && it.status !== "started";

    const cpuVal = parseFloat(it.cpu || "0") || 0;
    const gpuVal = parseFloat(it.gpu || "0") || 0;
    const memVal = parseFloat(it.mem_mb || "0") || 0;
    const netVal = parseFloat(it.net_kbs || "0") || 0;
    const ioVal = parseFloat(it.io_kbs || "0") || 0;
    const netUnitUpper = (data.net_unit || "KB").toUpperCase();
    const netIsMB = netUnitUpper === "MB";
    const netKBVal = netIsMB ? netVal * 1024 : netVal;
    const netMBVal = netIsMB ? netVal : netVal / 1024;
    const ioKBVal = netIsMB ? ioVal * 1024 : ioVal;
    const ioMBVal = netIsMB ? ioVal : ioVal / 1024;
    pushMetric(it.name, cpuVal, gpuVal, memVal, netVal, ioVal);
    const hist = metricHistory.get(it.name) || { cpu: [], gpu: [], mem: [], net: [], io: [] };
    const cpuSpark = buildSparkline(hist.cpu, "#67e8f9");
    const gpuSpark = buildSparkline(hist.gpu, "#fca5a5");
    const memSpark = buildSparkline(hist.mem, "#a7f3d0");
    const netSpark = buildSparkline(hist.net, "#c4b5fd");
    const ioSpark = buildSparkline(hist.io, "#f9d46b");

    const pidNum = Number(it.pid);
    const pidText = Number.isFinite(pidNum) && pidNum > 0
      ? `<span class="anim-number anim-pid">${Math.round(pidNum)}</span>`
      : "-";

    const netDisplay = netIsMB ? netVal.toFixed(2) : netVal.toFixed(0);
    const ioDisplay = netIsMB ? ioVal.toFixed(2) : ioVal.toFixed(0);

    tr.innerHTML = `
      <td>${it.name || ""}</td>
      <td>${it.type || ""}</td>
      <td class="status ${it.status || ""}">${it.icon || ""}</td>
      <td>${pidText}</td>
      <td>${it.started_at || "-"}</td>
      <td>${it.uptime || "-"}</td>
      <td class="metric">
        <div class="metric-wrap">
          <span class="metric-val anim-number anim-cpu">${cpuVal.toFixed(0)}%</span>
          ${cpuSpark}
        </div>
      </td>
      <td class="metric" title="GPU memory: ${it.gpu_mem_mb || 0} MB">
        <div class="metric-wrap">
          <span class="metric-val anim-number anim-gpu">${gpuVal.toFixed(0)}%</span>
          ${gpuSpark}
        </div>
      </td>
      <td class="metric" title="RAM: ${memVal.toFixed(2)} MB">
        <div class="metric-wrap">
          <span class="metric-val anim-number anim-ram">${memVal.toFixed(2)}MB</span>
          ${memSpark}
        </div>
      </td>
      <td class="metric" title="NET: ${netKBVal.toFixed(1)} KB/s | ${netMBVal.toFixed(2)} MB/s">
        <div class="metric-wrap">
          <span class="metric-val anim-number anim-net">${netDisplay}${netUnit}</span>
          ${netSpark}
        </div>
      </td>
      <td class="metric" title="IO: ${ioKBVal.toFixed(1)} KB/s | ${ioMBVal.toFixed(2)} MB/s">
        <div class="metric-wrap">
          <span class="metric-val anim-number anim-io">${ioDisplay}${netUnit}</span>
          ${ioSpark}
        </div>
      </td>
      <td>${it.target || ""}</td>
      <td>${it.error || ""}</td>
      <td>
        <button data-action="open-folder" data-name="${it.name}" title="Open folder">📁</button>
        <button data-action="start" data-name="${it.name}" ${canStart ? "" : "disabled"}>▶️</button>
        <button data-action="stop" data-name="${it.name}">❌</button>
        <button data-action="restart" data-name="${it.name}">🔄️</button>
      </td>
    `;
    tbody.appendChild(tr);

    const prevPid = Number(prev.pid);
    const prevCpu = parseFloat(prev.cpu || "0");
    const prevGpu = parseFloat(prev.gpu || "0");
    const prevMem = parseFloat(prev.mem_mb || "0");
    const prevNet = parseFloat(prev.net_kbs || "0");
    const prevIo = parseFloat(prev.io_kbs || "0");

    const pidEl = tr.querySelector(".anim-pid");
    if (pidEl && Number.isFinite(pidNum) && pidNum > 0) {
      animateNumber(
        pidEl,
        Number.isFinite(prevPid) && prevPid > 0 ? prevPid : pidNum,
        pidNum,
        (v) => `${Math.max(0, Math.round(v))}`
      );
    }
    animateNumber(
      tr.querySelector(".anim-cpu"),
      toFiniteOr(prevCpu, cpuVal),
      cpuVal,
      (v) => `${Math.max(0, Math.round(v))}%`
    );
    animateNumber(
      tr.querySelector(".anim-gpu"),
      toFiniteOr(prevGpu, gpuVal),
      gpuVal,
      (v) => `${Math.max(0, Math.round(v))}%`
    );
    animateNumber(
      tr.querySelector(".anim-ram"),
      toFiniteOr(prevMem, memVal),
      memVal,
      (v) => `${Math.max(0, v).toFixed(2)}MB`
    );
    animateNumber(
      tr.querySelector(".anim-net"),
      toFiniteOr(prevNet, netVal),
      netVal,
      (v) => netIsMB
        ? `${Math.max(0, v).toFixed(2)}${netUnit}`
        : `${Math.max(0, Math.round(v))}${netUnit}`
    );
    animateNumber(
      tr.querySelector(".anim-io"),
      toFiniteOr(prevIo, ioVal),
      ioVal,
      (v) => netIsMB
        ? `${Math.max(0, v).toFixed(2)}${netUnit}`
        : `${Math.max(0, Math.round(v))}${netUnit}`
    );
  }

  lastSnapshot = data;
};

const tick = async () => {
  if (!api) return;
  const data = await api.GetSnapshot();
  render(data);
};

tbody.addEventListener("click", async (e) => {
  const btn = e.target.closest("button");
  if (!btn || !api) return;
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
});

restartAllBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    await api.RestartAll();
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

killNodeBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    await api.KillNode();
  } catch (err) {
    console.error(err);
  }
});

const renderConfig = (model) => {
  if (!model) return;
  const s = model.settings || {};
  cfgCheckTiming.value = s.checkTiming || "";
  cfgRestartTiming.value = s.restartTiming || "";
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
};

const buildProcessRow = (p = {}) => {
  const card = document.createElement("div");
  card.className = "process-card";
  card.innerHTML = `
    <div class="process-grid">
      <label>Name
        <input data-f="name" value="${p.name || ""}" />
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
        <input data-f="process" value="${p.process || ""}" />
      </label>
      <label>Path
        <input data-f="path" value="${p.path || ""}" />
      </label>
      <label>Command
        <input data-f="command" value="${p.command || ""}" />
      </label>
      <label>Args
        <input data-f="args" value="${p.args || ""}" />
      </label>
      <label>CheckProcess
        <input data-f="checkProcess" value="${p.checkProcess || ""}" />
      </label>
      <label>CheckCmdline
        <input data-f="checkCmdline" value="${p.checkCmdline || ""}" />
      </label>
      <label>MonitorHang
        <input data-f="monitorHang" type="checkbox" ${p.monitorHang ? "checked" : ""} />
      </label>
      <label>HangTimeout
        <input data-f="hangTimeout" value="${p.hangTimeout || ""}" />
      </label>
    </div>
    <div class="process-actions">
      <button data-action="remove">Remove</button>
    </div>
  `;
  card.querySelector('select[data-f="type"]').value = p.type || "exe";
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
      checkProcess: get("checkProcess").value,
      checkCmdline: get("checkCmdline").value,
      monitorHang: get("monitorHang").checked,
      hangTimeout: get("hangTimeout").value,
    });
  }
  return {
    settings: {
      checkTiming: cfgCheckTiming.value,
      restartTiming: cfgRestartTiming.value,
      useETWNetwork: cfgUseETWNetwork.checked,
      netDebug: cfgNetDebug.checked,
      netUnit: cfgNetUnit.value,
      netScale: cfgNetScale.value,
      launchInNewConsole: cfgLaunchInNewConsole.checked,
      autoCloseErrorDialogs: cfgAutoCloseErrorDialogs.checked,
      errorWindowTitles: cfgErrorWindowTitles.value,
    },
    processes,
  };
};

reloadBtn.addEventListener("click", async () => {
  if (!api) return;
  if (!unlocked) return;
  const model = await api.GetConfigModel();
  renderConfig(model);
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
  lockConfig();
  openAuthModal();
});

setInterval(tick, 500);
window.onload = async () => {
  await tick();
  if (api) {
    if (unlocked) {
      const model = await api.GetConfigModel();
      renderConfig(model);
    }
  }
  lockConfig();
  configPanel.classList.add("hidden");
  document.querySelector(".config-actions").classList.add("hidden");
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

const lockConfig = () => {
  unlocked = false;
  document.querySelector(".config-grid").classList.add("hidden");
  document.querySelector(".process-list").classList.add("hidden");
  document.querySelector(".config-actions").classList.add("hidden");
  configPassword.value = "";
  configPanel.classList.add("hidden");
  configModal.classList.add("hidden");
};

const unlockConfig = async () => {
  if (configPassword.value !== PASSWORD) {
    alert("Неверный пароль");
    return;
  }
  unlocked = true;
  closeAuthModal();
  document.querySelector(".config-grid").classList.remove("hidden");
  document.querySelector(".process-list").classList.remove("hidden");
  document.querySelector(".config-actions").classList.remove("hidden");
  configPanel.classList.remove("hidden");
  configModal.classList.remove("hidden");
  const model = await api.GetConfigModel();
  renderConfig(model);
};

unlockBtn.addEventListener("click", unlockConfig);

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


// Checkbox pulse animation
document.addEventListener("change", (e) => {
  const el = e.target;
  if (el && el.type === "checkbox") {
    el.classList.remove("pulse");
    void el.offsetWidth;
    el.classList.add("pulse");
  }
});

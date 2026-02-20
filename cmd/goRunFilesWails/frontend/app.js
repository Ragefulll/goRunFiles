const api = window.go?.main?.GUI;

const elUpdated = document.getElementById("updated");
const elVersion = document.getElementById("version");
const tbody = document.getElementById("tbody");
const reloadBtn = document.getElementById("reloadConfig");
const saveBtn = document.getElementById("saveConfig");
const toggleBtn = document.getElementById("toggleConfig");
const addProcessBtn = document.getElementById("addProcess");
const configPanel = document.getElementById("configPanel");
const configPassword = document.getElementById("configPassword");
const unlockBtn = document.getElementById("unlockConfig");
const authModal = document.getElementById("authModal");
const closeAuth = document.getElementById("closeAuth");
const cancelAuth = document.getElementById("cancelAuth");

const cfgCheckTiming = document.getElementById("cfgCheckTiming");
const cfgRestartTiming = document.getElementById("cfgRestartTiming");
const cfgLaunchInNewConsole = document.getElementById("cfgLaunchInNewConsole");
const cfgAutoCloseErrorDialogs = document.getElementById("cfgAutoCloseErrorDialogs");
const cfgErrorWindowTitles = document.getElementById("cfgErrorWindowTitles");
const cfgProcesses = document.getElementById("configProcesses");

let lastSnapshot = null;

const render = (data) => {
  if (!data) return;
  elUpdated.textContent = data.updated || "—";
  elVersion.textContent = data.version || "—";
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

    tr.innerHTML = `
      <td>${it.name || ""}</td>
      <td>${it.type || ""}</td>
      <td class="status ${it.status || ""}">${it.icon || ""}</td>
      <td>${it.pid || "-"}</td>
      <td>${it.started_at || "-"}</td>
      <td>${it.uptime || "-"}</td>
      <td>${it.target || ""}</td>
      <td>${it.error || ""}</td>
      <td>
        <button data-action="start" data-name="${it.name}">▶️</button>
        <button data-action="stop" data-name="${it.name}">❌</button>
        <button data-action="restart" data-name="${it.name}">🔄️</button>
      </td>
    `;
    // no per-cell animations
    tbody.appendChild(tr);
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
    if (action === "start") await api.Start(name);
    if (action === "stop") await api.Stop(name);
    if (action === "restart") await api.Restart(name);
  } catch (err) {
    console.error(err);
  }
});

const renderConfig = (model) => {
  if (!model) return;
  const s = model.settings || {};
  cfgCheckTiming.value = s.checkTiming || "";
  cfgRestartTiming.value = s.restartTiming || "";
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
  if (configPanel.classList.contains("hidden")) {
    lockConfig();
    openAuthModal();
  } else {
    lockConfig();
  }
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


// Checkbox pulse animation
document.addEventListener("change", (e) => {
  const el = e.target;
  if (el && el.type === "checkbox") {
    el.classList.remove("pulse");
    void el.offsetWidth;
    el.classList.add("pulse");
  }
});

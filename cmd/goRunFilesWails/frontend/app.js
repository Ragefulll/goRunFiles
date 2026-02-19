const api = window.go?.main?.GUI;

const elUpdated = document.getElementById("updated");
const elVersion = document.getElementById("version");
const tbody = document.getElementById("tbody");
const reloadBtn = document.getElementById("reloadConfig");
const saveBtn = document.getElementById("saveConfig");
const toggleBtn = document.getElementById("toggleConfig");
const addProcessBtn = document.getElementById("addProcess");
const configPanel = document.getElementById("configPanel");

const cfgCheckTiming = document.getElementById("cfgCheckTiming");
const cfgRestartTiming = document.getElementById("cfgRestartTiming");
const cfgLaunchInNewConsole = document.getElementById("cfgLaunchInNewConsole");
const cfgAutoCloseErrorDialogs = document.getElementById("cfgAutoCloseErrorDialogs");
const cfgErrorWindowTitles = document.getElementById("cfgErrorWindowTitles");
const cfgProcesses = document.getElementById("configProcesses");

const render = (data) => {
  if (!data) return;
  elUpdated.textContent = data.updated || "—";
  elVersion.textContent = data.version || "—";
  tbody.innerHTML = "";

  for (const it of data.items || []) {
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
        <button data-action="start" data-name="${it.name}">Start</button>
        <button data-action="stop" data-name="${it.name}">Stop</button>
        <button data-action="restart" data-name="${it.name}">Restart</button>
      </td>
    `;
    tbody.appendChild(tr);
  }
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
  const tr = document.createElement("tr");
  tr.innerHTML = `
    <td><input data-f="name" value="${p.name || ""}" /></td>
    <td><input data-f="disabled" type="checkbox" ${p.disabled ? "checked" : ""} /></td>
    <td>
      <select data-f="type">
        <option value="exe">exe</option>
        <option value="cmd">cmd</option>
        <option value="bat">bat</option>
      </select>
    </td>
    <td><input data-f="process" value="${p.process || ""}" /></td>
    <td><input data-f="path" value="${p.path || ""}" /></td>
    <td><input data-f="command" value="${p.command || ""}" /></td>
    <td><input data-f="args" value="${p.args || ""}" /></td>
    <td><input data-f="checkProcess" value="${p.checkProcess || ""}" /></td>
    <td><input data-f="checkCmdline" value="${p.checkCmdline || ""}" /></td>
    <td><input data-f="monitorHang" type="checkbox" ${p.monitorHang ? "checked" : ""} /></td>
    <td><input data-f="hangTimeout" value="${p.hangTimeout || ""}" /></td>
    <td><button data-action="remove">Remove</button></td>
  `;
  tr.querySelector('select[data-f="type"]').value = p.type || "exe";
  return tr;
};

const collectConfig = () => {
  const processes = [];
  const names = new Set();
  for (const tr of cfgProcesses.querySelectorAll("tr")) {
    const get = (f) => tr.querySelector(`[data-f="${f}"]`);
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
  const model = await api.GetConfigModel();
  renderConfig(model);
});

saveBtn.addEventListener("click", async () => {
  if (!api) return;
  try {
    const model = collectConfig();
    await api.SaveConfigModel(model);
  } catch (err) {
    alert(err.message || String(err));
  }
});

toggleBtn.addEventListener("click", () => {
  configPanel.classList.toggle("hidden");
  document.querySelector(".config-actions").classList.toggle("hidden");
});

setInterval(tick, 500);
window.onload = async () => {
  await tick();
  if (api) {
    const model = await api.GetConfigModel();
    renderConfig(model);
  }
  configPanel.classList.add("hidden");
  document.querySelector(".config-actions").classList.add("hidden");
};

addProcessBtn.addEventListener("click", () => {
  cfgProcesses.appendChild(buildProcessRow({}));
});

cfgProcesses.addEventListener("click", (e) => {
  const btn = e.target.closest("button");
  if (!btn) return;
  if (btn.dataset.action === "remove") {
    btn.closest("tr").remove();
  }
});

import { gui } from "./api.js";
import { clone } from "./util.js";

const { reactive } = window.Vue;

const HISTORY_LEN = 40;
const MIN_TICK_MS = 100;
const ERROR_LOG_MAX = 600;
const CMD_CHECK_CMDLINE_EXCLUDE_DEFAULT = "jetbrains,js-language-service,typingsinstaller,eslint";
const PASSWORD = "art3d";
const UNLOCK_KEY = "goRunFilesUnlocked";

export const store = reactive({
  // header
  windowLabel: "",
  version: "—",
  updated: "—",
  netStatus: "—",
  netStatusTitle: "",
  netDebug: "—",

  // process monitoring
  snapshot: null,
  items: [],
  history: {},
  checkProcessRunning: true,
  checkToggleBusy: false,

  // error console
  consoleOpened: false,
  errorLog: [],

  // auth / config access
  unlocked: false,
  password: "",
  authError: "",
  authModalOpen: false,
  pendingOpenConfig: false,
  pendingProcessOpen: "",
  toggleBusy: {},

  // configuration
  configModel: null,
  configModalOpen: false,
  screens: [],
  screensError: "",
  cfgFind: "",
  activeProcessName: "",

  // drawer
  drawerOpen: false,
  drawerName: "",
  drawerProcess: null,
  drawerBusy: false,

  // scheduler
  scheduler: null,
  schedulerBusy: false,

  // auto update popup
  update: {
    visible: false,
    icon: "",
    title: "",
    msg: "",
    detail: "",
    pct: "",
    cls: "",
    indeterminate: false,
  },
});

const pushMetric = (name, cpu, gpu, mem, net, io) => {
  let h = store.history[name];
  if (!h) h = store.history[name] = { cpu: [], gpu: [], mem: [], net: [], io: [] };
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

const formatStamp = (iso) => (iso || new Date().toISOString().replace("T", " ").slice(0, 19));

const lastErrorByProcess = new Map();

const collectErrorLog = (data) => {
  const stamp = formatStamp(data?.updated);
  const activeNames = new Set();
  for (const it of data.items || []) {
    const name = it.name || "unknown";
    activeNames.add(name);
    const currentError = (it.error || "").trim();
    const prevError = lastErrorByProcess.get(name) || "";
    if (currentError && currentError !== prevError) {
      store.errorLog.unshift(`[${stamp}] ${name}: ${currentError}`);
    }
    lastErrorByProcess.set(name, currentError);
  }
  for (const name of Array.from(lastErrorByProcess.keys())) {
    if (!activeNames.has(name)) lastErrorByProcess.delete(name);
  }
  if (store.errorLog.length > ERROR_LOG_MAX) {
    store.errorLog.splice(ERROR_LOG_MAX);
  }
};

export const getSnapshot = async () => {
  if (!gui) return;
  const data = await gui.GetSnapshot();
  applySnapshot(data);
};

const applySnapshot = (data) => {
  if (!data) return;
  store.snapshot = data;
  store.items = data.items || [];
  store.updated = data.updated || "—";
  if (data.version) store.version = data.version;
  store.checkProcessRunning = data.check_process_running !== false;

  const mode = data.net_mode || "—";
  const err = data.net_err || "";
  store.netStatus = err ? `${mode} (${err})` : mode;
  store.netStatusTitle = err || "";
  store.netDebug = data.net_dbg || "—";

  collectErrorLog(data);

  for (const it of store.items) {
    const name = it.name || "";
    if (!name) continue;
    pushMetric(
      name,
      parseFloat(it.cpu || "0") || 0,
      parseFloat(it.gpu || "0") || 0,
      parseFloat(it.mem_mb || "0") || 0,
      parseFloat(it.net_kbs || "0") || 0,
      parseFloat(it.io_kbs || "0") || 0
    );
  }
};

// ---- polling ----
let tickTimer = null;
let tickIntervalMs = 500;
let tickInFlight = false;

const scheduleTick = (delayMs) => {
  if (tickTimer) clearTimeout(tickTimer);
  tickTimer = setTimeout(runTick, delayMs);
};

const runTick = async () => {
  if (tickInFlight || !gui) {
    scheduleTick(tickIntervalMs);
    return;
  }
  tickInFlight = true;
  const start = performance.now();
  try {
    await getSnapshot();
  } finally {
    tickInFlight = false;
    const data = store.snapshot;
    if (data && Number.isFinite(data.check_timing_ms) && data.check_timing_ms > 0) {
      tickIntervalMs = Math.max(MIN_TICK_MS, data.check_timing_ms);
    }
    const elapsed = performance.now() - start;
    scheduleTick(Math.max(0, tickIntervalMs - elapsed));
  }
};

export const startPolling = () => scheduleTick(tickIntervalMs);

// ---- screens ----
export const refreshScreens = async () => {
  if (!gui?.GetScreens) {
    store.screens = [];
    store.screensError = "Screen list is unavailable";
    return;
  }
  try {
    const screens = await gui.GetScreens();
    store.screens = Array.isArray(screens) ? screens : [];
    store.screensError = "";
  } catch (err) {
    store.screens = [];
    store.screensError = err.message || String(err);
  }
};

// ---- config model ----
const setConfigModel = (model) => {
  store.configModel = clone(model || { settings: {}, processes: [] });
  if (!store.configModel.settings) store.configModel.settings = {};
  if (!Array.isArray(store.configModel.processes)) store.configModel.processes = [];
};

const findProcess = (name) => {
  const model = store.configModel;
  if (!model || !Array.isArray(model.processes)) return null;
  return model.processes.find((p) => p.name === name) || null;
};

export { setConfigModel, findProcess, CMD_CHECK_CMDLINE_EXCLUDE_DEFAULT };

// ---- auth / unlock ----
export const isConfigUnlocked = () =>
  store.unlocked || localStorage.getItem(UNLOCK_KEY) === "1";

const setUnlocked = (value) => {
  store.unlocked = !!value;
  localStorage.setItem(UNLOCK_KEY, store.unlocked ? "1" : "0");
};

export const openAuthModal = () => {
  store.authModalOpen = true;
  store.authError = "";
  store.password = "";
};

export const closeAuthModal = () => {
  store.authModalOpen = false;
  store.password = "";
};

export const unlockConfig = async () => {
  if (store.password !== PASSWORD) {
    store.authError = "Неверный пароль";
    return;
  }
  setUnlocked(true);
  closeAuthModal();
  if (store.pendingOpenConfig) {
    store.pendingOpenConfig = false;
    await openConfigModal();
  }
  if (store.pendingProcessOpen) {
    const name = store.pendingProcessOpen;
    store.pendingProcessOpen = "";
    if (!store.configModel && gui) {
      const model = await gui.GetConfigModel();
      setConfigModel(model);
    }
    openDrawer(name);
  }
  await refreshSchedulerStatus();
};

export const lockConfig = () => {
  setUnlocked(false);
  store.pendingOpenConfig = false;
  store.pendingProcessOpen = "";
  store.password = "";
  store.cfgFind = "";
  store.configModalOpen = false;
  closeDrawer();
  store.activeProcessName = "";
};

export const openConfigModal = async () => {
  if (!isConfigUnlocked()) {
    store.pendingOpenConfig = true;
    openAuthModal();
    return;
  }
  store.configModalOpen = true;
  await refreshScreens();
  if (!store.configModel && gui) {
    const model = await gui.GetConfigModel();
    setConfigModel(model);
  }
};

export const toggleConfigOpen = () => openConfigModal();

export const reloadConfig = async () => {
  if (!gui || !store.unlocked) return;
  await refreshScreens();
  const model = await gui.GetConfigModel();
  setConfigModel(model);
};

export const saveConfig = async () => {
  if (!gui || !store.configModel) return;
  const model = clone(store.configModel);
  try {
    model.processes = Array.isArray(model.processes) ? model.processes : [];
    const names = new Set();
    for (const p of model.processes) {
      const name = (p.name || "").trim();
      if (!name) throw new Error("Process name is required");
      if (names.has(name)) throw new Error(`Duplicate process name: ${name}`);
      names.add(name);
      if (p.type !== "cmd") p.command = "";
    }
    await gui.SaveConfigModel(model);
    setConfigModel(model);
  } catch (err) {
    alert(err.message || String(err));
    throw err;
  }
};

export const addProcess = () => {
  if (!store.configModel || !Array.isArray(store.configModel.processes)) return;
  store.configModel.processes.push({
    name: "",
    disabled: false,
    type: "exe",
    process: "",
    path: "",
    command: "",
    args: "",
    screen: 0,
    checkProcess: "",
    checkCmdline: "",
    checkCmdlineExclude: "",
    delayStartTime: "",
    monitorHang: false,
    hangTimeout: "",
  });
};

export const removeProcess = (index) => {
  if (!store.configModel || !Array.isArray(store.configModel.processes)) return;
  store.configModel.processes.splice(index, 1);
};

// ---- process card controls ----
export const processAction = async (name, action) => {
  if (!gui) return;
  try {
    if (action === "open-folder") await gui.OpenFolder(name);
    else if (action === "start") await gui.Start(name);
    else if (action === "stop") await gui.Stop(name);
    else if (action === "restart") await gui.Restart(name);
  } catch (err) {
    console.error(err);
  }
};

export const toggleDisabled = async (item) => {
  if (!gui) return;
  const name = item.name;
  const target = !item.disabled;
  const prev = item.disabled;
  item.disabled = target;
  store.toggleBusy[name] = true;
  try {
    const model = await gui.GetConfigModel();
    if (!model || !Array.isArray(model.processes)) throw new Error("Config load failed");
    const p = model.processes.find((x) => x.name === name);
    if (p) p.disabled = target;
    await gui.SaveConfigModel(model);
    setConfigModel(model);
    await getSnapshot();
  } catch (err) {
    item.disabled = prev;
    alert(err.message || String(err));
  } finally {
    store.toggleBusy[name] = false;
  }
};

export const toggleCheckProcess = async () => {
  if (!gui || store.checkToggleBusy) return;
  store.checkToggleBusy = true;
  try {
    const running = store.checkProcessRunning
      ? await gui.StopCheckProcess()
      : await gui.StartCheckProcess();
    store.checkProcessRunning = !!running;
    await getSnapshot();
  } catch (err) {
    console.error(err);
  } finally {
    store.checkToggleBusy = false;
  }
};

export const restartAll = async () => {
  if (!gui) return;
  try {
    await gui.RestartAll();
  } catch (err) {
    console.error(err);
  }
};

export const restartAutoManual = async () => {
  if (!gui) return;
  try {
    await gui.RestartAutoManual();
  } catch (err) {
    console.error(err);
  }
};

export const stopAll = async () => {
  if (!gui) return;
  try {
    await gui.StopAll();
  } catch (err) {
    console.error(err);
  }
};

export const killCMD = async () => {
  if (!gui) return;
  try {
    await gui.KillCMD();
  } catch (err) {
    console.error(err);
  }
};

export const killNode = async () => {
  if (!gui) return;
  try {
    await gui.KillNode();
  } catch (err) {
    console.error(err);
  }
};

// ---- drawer ----
export const openProcessEditor = async (name) => {
  if (!gui) return;
  store.pendingProcessOpen = name;
  if (!isConfigUnlocked()) {
    openAuthModal();
    return;
  }
  await refreshScreens();
  const model = await gui.GetConfigModel();
  setConfigModel(model);
  openDrawer(name);
};

export const openDrawer = (name) => {
  const p = findProcess(name);
  if (!p) return;
  store.drawerOpen = true;
  store.drawerName = name;
  store.drawerProcess = clone(p);
  store.activeProcessName = name;
};

export const closeDrawer = () => {
  store.drawerOpen = false;
  store.drawerName = "";
  store.drawerProcess = null;
};

export const saveDrawer = async () => {
  if (!gui || !store.configModel) return;
  const draft = clone(store.drawerProcess || {});
  if (!(draft.name || "").trim()) throw new Error("Name is required");
  const next = clone(store.configModel);
  next.processes = Array.isArray(next.processes) ? next.processes : [];
  const idx = next.processes.findIndex((p) => p.name === store.drawerName);
  if (idx < 0) throw new Error("Process not found");
  next.processes[idx] = draft;
  await gui.SaveConfigModel(next);
  setConfigModel(next);
  await refreshScreens();
  openDrawer(draft.name);
  await getSnapshot();
};

// ---- error console ----
export const toggleConsole = () => {
  store.consoleOpened = !store.consoleOpened;
};

// ---- scheduler ----
export const refreshSchedulerStatus = async () => {
  if (!gui) {
    store.scheduler = null;
    return;
  }
  try {
    const s = await gui.GetSchedulerStatus();
    store.scheduler = s;
  } catch (err) {
    store.scheduler = { taskName: "—", error: err.message || String(err) };
  }
};

export const installScheduler = async () => {
  if (!gui || store.schedulerBusy) return;
  store.schedulerBusy = true;
  try {
    await gui.InstallScheduler();
    await refreshSchedulerStatus();
  } catch (err) {
    alert(err.message || String(err));
  } finally {
    store.schedulerBusy = false;
  }
};

export const removeScheduler = async () => {
  if (!gui || store.schedulerBusy) return;
  store.schedulerBusy = true;
  try {
    await gui.RemoveScheduler();
    await refreshSchedulerStatus();
  } catch (err) {
    alert(err.message || String(err));
  } finally {
    store.schedulerBusy = false;
  }
};

// ---- auto update ----
const UPDATE_ICONS = {
  spinner:
    '<svg class="spin" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-dasharray="40 20" /></svg>',
  download:
    '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 4v10m0 0-4-4m4 4 4-4M5 17v1a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-1" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" /></svg>',
  check:
    '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 6 9 17l-5-5" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" /></svg>',
  alert:
    '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 8v5m0 3v.01M12 3 2.5 20h19L12 3z" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" /></svg>',
};

let updateTimer = null;

const showUpdateCard = (title, msg, detail, pct, cls, ico, indeterminate) => {
  store.update.visible = true;
  store.update.cls = cls || "";
  store.update.indeterminate = !!indeterminate;
  store.update.icon = UPDATE_ICONS[ico] || "";
  store.update.title = title;
  store.update.msg = msg;
  store.update.detail = detail || "";
  store.update.pct = pct || "";
};

const hideUpdate = () => {
  store.update.visible = false;
};

export const onUpdateStatus = (ev) => {
  if (!ev || !ev.status) return;
  if (updateTimer) {
    clearTimeout(updateTimer);
    updateTimer = null;
  }
  const cur = ev.current ? `v${ev.current}` : "-";
  switch (ev.status) {
    case "check":
      showUpdateCard("Проверка обновлений", `Текущая версия: ${cur}`, "Подключение к серверу...", "", "", "spinner", true);
      break;
    case "downloading":
      showUpdateCard(`Доступно обновление`, `Загружается версия v${ev.remote || ""}`, ev.detail || "", String(Math.round(ev.progress || 0)), "downloading", "download");
      break;
    case "restarting":
      showUpdateCard("Установка обновления", "Файлы заменены, приложение перезапускается...", "", "100%", "restarting", "spinner");
      break;
    case "applied":
      showUpdateCard("Обновление применено", `Установлена версия v${ev.current || ""}`, "", "100%", "ok", "check");
      updateTimer = setTimeout(hideUpdate, 2000);
      break;
    case "idle":
      showUpdateCard("Версия актуальна", `Установлена актуальная версия v${ev.current || ""}`, "", "100%", "ok", "check");
      updateTimer = setTimeout(hideUpdate, 2000);
      break;
    default:
      showUpdateCard("Ошибка", ev.detail || "Попробуйте позже", "", "100%", "error", "alert");
      updateTimer = setTimeout(hideUpdate, 4500);
  }
};

// ---- boot ----
const start = async () => {
  store.unlocked = localStorage.getItem(UNLOCK_KEY) === "1";

  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("update-status", onUpdateStatus);
  }

  if (gui?.AppVersion) {
    try {
      const v = await gui.AppVersion();
      if (v) store.version = v;
    } catch {
      /* ignore */
    }
  }

  if (gui) {
    await getSnapshot();
    if (store.unlocked) {
      await refreshScreens();
      const model = await gui.GetConfigModel();
      setConfigModel(model);
    }
  }

  await refreshSchedulerStatus();

  if (gui?.CheckUpdates) gui.CheckUpdates();
  if (gui?.GetUpdateStatus) {
    try {
      const st = await gui.GetUpdateStatus();
      if (st && st.status && !["check", "downloading", "restarting"].includes(st.status)) {
        onUpdateStatus(st);
      }
    } catch {
      /* ignore */
    }
  }

  startPolling();
};

Object.assign(store, {
  start,
  getSnapshot,
  refreshScreens,
  isConfigUnlocked,
  openAuthModal,
  closeAuthModal,
  unlockConfig,
  lockConfig,
  openConfigModal,
  toggleConfigOpen,
  reloadConfig,
  saveConfig,
  addProcess,
  removeProcess,
  processAction,
  toggleDisabled,
  toggleCheckProcess,
  restartAll,
  restartAutoManual,
  stopAll,
  killCMD,
  killNode,
  openProcessEditor,
  openDrawer,
  closeDrawer,
  saveDrawer,
  toggleConsole,
  refreshSchedulerStatus,
  installScheduler,
  removeScheduler,
  onUpdateStatus,
});
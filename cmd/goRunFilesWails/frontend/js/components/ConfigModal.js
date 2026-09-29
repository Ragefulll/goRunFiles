import { store, CMD_CHECK_CMDLINE_EXCLUDE_DEFAULT } from "../store.js";
import { ScreenViz, MonitorPicker } from "./MonitorPicker.js";

const DEFAULT_MODEL = { settings: {}, processes: [] };

const SchedulerBlock = {
  name: "SchedulerBlock",
  computed: {
    store() {
      return store;
    },
    s() {
      return store.scheduler || null;
    },
    taskName() {
      return this.s?.taskName || "—";
    },
    stateText() {
      if (!this.s) return "—";
      if (!this.s.installed) return "Not installed";
      return this.s.running ? "Running" : this.s.state || "Installed";
    },
    lastRun() {
      return this.s?.lastRunTime || "—";
    },
    lastResult() {
      return Number.isFinite(this.s?.lastTaskResult) ? String(this.s.lastTaskResult) : "—";
    },
    note() {
      if (!this.s) return "—";
      const parts = [];
      if (this.s.installed) parts.push("Watchdog restarts app on exit.");
      else parts.push("Install to auto-start with admin rights.");
      if (this.s.error) parts.push(this.s.error);
      return parts.join(" ");
    },
  },
  template: `
    <div class="scheduler">
      <div class="scheduler-status">
        <div>Task: <span>{{ taskName }}</span></div>
        <div>Status: <span>{{ stateText }}</span></div>
        <div>Last run: <span>{{ lastRun }}</span></div>
        <div>Last result: <span>{{ lastResult }}</span></div>
        <div class="scheduler-note" id="schedulerNote">{{ note }}</div>
      </div>
      <div class="scheduler-actions">
        <button class="panel-actions__button fixed" :disabled="store.schedulerBusy" @click="store.installScheduler()">Install</button>
        <button class="panel-actions__button fixed" :disabled="store.schedulerBusy" @click="store.removeScheduler()">Remove</button>
        <button class="panel-actions__button fixed" :disabled="store.schedulerBusy" @click="store.refreshSchedulerStatus()">Refresh</button>
      </div>
    </div>
  `,
};

const ConfigProcessRow = {
  name: "ConfigProcessRow",
  components: { MonitorPicker },
  props: {
    process: { type: Object, required: true },
    index: { type: Number, required: true },
    hidden: { type: Boolean, default: false },
  },
  emits: ["remove"],
  computed: {
    store() {
      return store;
    },
    isCmd() {
      return this.process.type === "cmd";
    },
    labelStyle() {
      return { opacity: this.isCmd ? "1" : "0.3" };
    },
    rowClasses() {
      return {
        "hidden-by-filter": this.hidden,
        selected: store.activeProcessName === this.process.name && !!this.process.name,
      };
    },
  },
  watch: {
    "process.type"(val) {
      if (val === "cmd") {
        if (!(this.process.checkCmdlineExclude || "").trim()) {
          this.process.checkCmdlineExclude = CMD_CHECK_CMDLINE_EXCLUDE_DEFAULT;
        }
      } else {
        this.process.command = "";
      }
    },
  },
  methods: {
    onFocusName() {
      if (this.process.name) store.activeProcessName = this.process.name;
    },
  },
  template: `
    <div class="process-card" :class="rowClasses" :data-name="process.name">
      <div class="process-grid">
        <label>Disabled
          <input type="checkbox" v-model="process.disabled" />
        </label>
        <label>Name
          <input v-model="process.name" @focusin="onFocusName" />
        </label>
        <label>Screen
          <MonitorPicker v-model="process.screen" />
        </label>
        <label>Type
          <select v-model="process.type">
            <option value="exe">exe</option>
            <option value="cmd">cmd</option>
            <option value="bat">bat</option>
          </select>
        </label>
        <label>Process
          <input v-model="process.process" />
        </label>
        <label>Path
          <input v-model="process.path" />
        </label>
        <label :style="labelStyle">CMD Command (npm run start и тд)
          <input v-model="process.command" :disabled="!isCmd" />
        </label>
        <label>Args
          <input v-model="process.args" />
        </label>
        <label>CheckProcess (PROJECT.exe, PROJECT-Win64-Shipping.exe || node.exe)
          <input v-model="process.checkProcess" />
        </label>
        <label>CheckCmdline (name=PC2 || name=PC1 || ue-project.art3d.loc nuxt)
          <input v-model="process.checkCmdline" />
        </label>
        <label>CheckCmdlineExclude (jetbrains,js-language-service,typingsinstaller,eslint)
          <input v-model="process.checkCmdlineExclude" />
        </label>
        <label>DelayStartTime
          <input v-model="process.delayStartTime" placeholder="30s" />
        </label>
        <label>MonitorHang
          <input v-model="process.monitorHang" type="checkbox" />
        </label>
        <label>HangTimeout
          <input v-model="process.hangTimeout" />
        </label>
      </div>
      <div class="process-actions">
        <button class="panel-actions__button" @click="$emit('remove')">Remove</button>
      </div>
    </div>
  `,
};

const ConfigModal = {
  name: "ConfigModal",
  components: { SchedulerBlock, ScreenViz, ConfigProcessRow },
  computed: {
    store() {
      return store;
    },
    model() {
      return store.configModel || DEFAULT_MODEL;
    },
    settings() {
      return this.model.settings || {};
    },
    processes() {
      return Array.isArray(this.model.processes) ? this.model.processes : [];
    },
  },
  methods: {
    filterMatch(p) {
      const q = (store.cfgFind || "").trim().toLowerCase();
      if (!q) return false;
      return !((p.name || "").trim().toLowerCase().includes(q));
    },
    clearFind() {
      store.cfgFind = "";
    },
    async onSave() {
      try {
        await store.saveConfig();
      } catch {
        /* error already reported in store */
      }
    },
  },
  template: `
    <div class="modal" v-if="store.configModalOpen">
      <div class="modal-backdrop" @click="store.lockConfig()"></div>
      <div class="modal-card modal-wide">
        <div class="modal-head">
          <div>Config</div>
        </div>
        <div class="config-actions">
          <button class="panel-actions__button fixed" @click="store.reloadConfig()">Reload</button>
          <button class="panel-actions__button fixed" @click="onSave">Save</button>
          <button class="panel-actions__button fixed" @click="store.lockConfig()">Lock</button>
          <button class="panel-actions__button fixed" @click="store.addProcess()">Add process</button>
          <button title="Закрыть" @click="store.lockConfig()">✕</button>
        </div>
        <div class="config-panel" id="configPanel">
          <h3>Scheduler</h3>
          <SchedulerBlock />
          <h3>Screens</h3>
          <ScreenViz />
          <div class="config-grid">
            <label>Check timing
              <input v-model="settings.checkTiming" placeholder="500ms" />
            </label>
            <label>Restart timing
              <input v-model="settings.restartTiming" placeholder="3s" />
            </label>
            <label>Auto restart
              <input v-model="settings.autoRestart" type="checkbox" />
            </label>
            <label>Auto restart time
              <input v-model="settings.autoRestartTime" type="time" step="60" placeholder="04:00" />
            </label>
            <label>Restart on exit (scheduler)
              <input v-model="settings.autoRestartOnExit" type="checkbox" />
            </label>
            <label>Use ETW network
              <input v-model="settings.useETWNetwork" type="checkbox" />
            </label>
            <label>Net debug
              <input v-model="settings.netDebug" type="checkbox" />
            </label>
            <label>Net unit
              <select v-model="settings.netUnit">
                <option value="KB">KB/s</option>
                <option value="MB">MB/s</option>
              </select>
            </label>
            <label>Net scale
              <select v-model="settings.netScale">
                <option value="1">1</option>
                <option value="100">/100</option>
                <option value="1000">/1000</option>
              </select>
            </label>
            <label>Launch in new console
              <input v-model="settings.launchInNewConsole" type="checkbox" />
            </label>
            <label>Auto close error dialogs
              <input v-model="settings.autoCloseErrorDialogs" type="checkbox" />
            </label>
          </div>
          <label class="full">Error window titles
            <input v-model="settings.errorWindowTitles" />
          </label>
          <label class="full">Find
            <input v-model="store.cfgFind" @keydown.escape="clearFind" />
          </label>
          <h3>Processes</h3>
          <div class="process-list">
            <ConfigProcessRow
              v-for="(p, i) in processes"
              :key="i"
              :process="p"
              :index="i"
              :hidden="filterMatch(p)"
              @remove="store.removeProcess(i)"
            />
          </div>
        </div>
      </div>
    </div>
  `,
};

export { ConfigModal };
import { store } from "../store.js";
import { statusInfo } from "../util.js";
import { Sparkline, AnimatedNumber } from "./units.js";
import { MonitorPicker } from "./MonitorPicker.js";

const EMPTY_HIST = { cpu: [], gpu: [], mem: [], net: [], io: [] };

const ProcessDrawer = {
  name: "ProcessDrawer",
  components: { Sparkline, AnimatedNumber, MonitorPicker },
  computed: {
    store() {
      return store;
    },
    draft() {
      return store.drawerProcess || {};
    },
    p() {
      return store.drawerProcess || {};
    },
    item() {
      if (!store.drawerName) return null;
      return store.items.find((it) => it.name === store.drawerName) || null;
    },
    status() {
      return statusInfo(this.item || this.p);
    },
    statusText() {
      return this.p.type
        ? `${this.p.type.toUpperCase()} \u2022 ${this.p.path || this.p.process || ""}`
        : this.p.path || this.p.process || "\u2014";
    },
    hist() {
      return store.drawerName ? store.history[store.drawerName] || EMPTY_HIST : EMPTY_HIST;
    },
    cpuVal() {
      return this.item ? parseFloat(this.item.cpu || "0") || 0 : null;
    },
    gpuVal() {
      return this.item ? parseFloat(this.item.gpu || "0") || 0 : null;
    },
    memVal() {
      return this.item ? parseFloat(this.item.mem_mb || "0") || 0 : null;
    },
    netVal() {
      return this.item ? parseFloat(this.item.net_kbs || "0") || 0 : null;
    },
    ioVal() {
      return this.item ? parseFloat(this.item.io_kbs || "0") || 0 : null;
    },
    pctFmt() {
      return (v) => `${Math.max(0, Math.round(v))}%`;
    },
    memFmt() {
      return (v) => `${Math.max(0, v).toFixed(2)}MB`;
    },
    kbFmt() {
      return (v) => `${Math.max(0, Math.round(v))}KB/s`;
    },
  },
  methods: {
    act(action) {
      if (store.drawerName) store.processAction(store.drawerName, action);
    },
    async save() {
      try {
        await store.saveDrawer();
      } catch (err) {
        alert(err.message || String(err));
      }
    },
  },
  template: `
    <aside class="drawer" :aria-hidden="!store.drawerOpen" v-if="store.drawerOpen">
      <div class="drawer-backdrop" @click="store.closeDrawer()"></div>
      <div class="drawer-panel">
        <div class="drawer-head">
          <div>
            <div class="drawer-kicker">Process detail</div>
            <div class="drawer-title">{{ p.name || '—' }}</div>
          </div>
          <button title="Close" @click="store.closeDrawer()">✕</button>
        </div>
        <div class="drawer-status">
          <span class="drawer-chip" :class="status.cls"><i class="mdl2 status-ico">{{ status.icon }}</i> {{ status.label }}</span>
          <span>{{ statusText }}</span>
        </div>
        <div class="drawer-actions">
          <button class="neon-btn neon-btn--icon neon-btn--folder" title="Open folder" aria-label="Open folder" @click="act('open-folder')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6.5A2.5 2.5 0 0 1 5.5 4h4.2a2 2 0 0 1 1.5.7l1 1.3H18.5A2.5 2.5 0 0 1 21 8.5v7A2.5 2.5 0 0 1 18.5 18h-13A2.5 2.5 0 0 1 3 15.5v-9Z"/></svg>
          </button>
          <button class="neon-btn neon-btn--icon neon-btn--restart" title="Restart" aria-label="Restart" @click="act('restart')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5a7 7 0 1 1-6.1 10.4M7 5H3v4"/><path d="M3 9c1.4-3.4 4.7-5.8 8.5-5.8A8.5 8.5 0 1 1 5.2 18.5"/></svg>
          </button>
          <button class="neon-btn neon-btn--icon neon-btn--stop" title="Stop" aria-label="Stop" @click="act('stop')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="6.5" y="6.5" width="11" height="11" rx="2.2"/></svg>
          </button>
          <button class="neon-btn neon-btn--icon neon-btn--start" title="Start" aria-label="Start" @click="act('start')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5.5v13l11-6.5-11-6.5Z"/></svg>
          </button>
        </div>
        <div class="drawer-stats">
          <div><span>PID</span><strong>{{ item ? item.pid || '—' : '—' }}</strong></div>
          <div><span>Started</span><strong>{{ item ? item.started_at || '—' : '—' }}</strong></div>
          <div><span>Uptime</span><strong>{{ item ? item.uptime || '—' : '—' }}</strong></div>
          <div class="full"><span>Target</span><strong>{{ item ? item.target || '—' : '—' }}</strong></div>
        </div>
        <div class="drawer-metrics">
          <div class="drawer-metric">
            <span>CPU</span>
            <template v-if="cpuVal !== null"><AnimatedNumber :value="cpuVal" :format="pctFmt" /></template>
            <strong v-else>—</strong>
            <div class="spark-wrap"><Sparkline :values="hist.cpu" color="#67e8f9" /></div>
          </div>
          <div class="drawer-metric">
            <span>GPU</span>
            <template v-if="gpuVal !== null"><AnimatedNumber :value="gpuVal" :format="pctFmt" /></template>
            <strong v-else>—</strong>
            <div class="spark-wrap"><Sparkline :values="hist.gpu" color="#fca5a5" /></div>
          </div>
          <div class="drawer-metric">
            <span>RAM</span>
            <template v-if="memVal !== null"><AnimatedNumber :value="memVal" :format="memFmt" /></template>
            <strong v-else>—</strong>
            <div class="spark-wrap"><Sparkline :values="hist.mem" color="#a7f3d0" /></div>
          </div>
          <div class="drawer-metric">
            <span>NET</span>
            <template v-if="netVal !== null"><AnimatedNumber :value="netVal" :format="kbFmt" /></template>
            <strong v-else>—</strong>
            <div class="spark-wrap"><Sparkline :values="hist.net" color="#c4b5fd" /></div>
          </div>
          <div class="drawer-metric">
            <span>IO</span>
            <template v-if="ioVal !== null"><AnimatedNumber :value="ioVal" :format="kbFmt" /></template>
            <strong v-else>—</strong>
            <div class="spark-wrap"><Sparkline :values="hist.io" color="#f9d46b" /></div>
          </div>
        </div>
        <div class="drawer-form">
          <label>Name <input v-model="p.name" /></label>
          <label>Disabled <input v-model="p.disabled" type="checkbox" /></label>
          <label>Type
            <select v-model="p.type">
              <option value="exe">exe</option>
              <option value="cmd">cmd</option>
              <option value="bat">bat</option>
            </select>
          </label>
          <label>Process <input v-model="p.process" /></label>
          <label>Path <input v-model="p.path" /></label>
          <label>Command <input v-model="p.command" /></label>
          <label>Args <input v-model="p.args" /></label>
          <label>Screen <input v-model.number="p.screen" type="hidden" /><MonitorPicker v-model="p.screen" /></label>
          <label>CheckProcess <input v-model="p.checkProcess" /></label>
          <label>CheckCmdline <input v-model="p.checkCmdline" /></label>
          <label>CheckCmdlineExclude <input v-model="p.checkCmdlineExclude" /></label>
          <label>DelayStartTime <input v-model="p.delayStartTime" placeholder="30s" /></label>
          <label>MonitorHang <input v-model="p.monitorHang" type="checkbox" /></label>
          <label>HangTimeout <input v-model="p.hangTimeout" /></label>
        </div>
        <div class="drawer-footer">
          <button class="neon-btn neon-btn--restart" @click="save">Save</button>
        </div>
      </div>
    </aside>
  `,
};

export { ProcessDrawer };
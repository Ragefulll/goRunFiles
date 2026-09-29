import { store } from "../store.js";
import { statusInfo } from "../util.js";
import { Sparkline, AnimatedNumber } from "./units.js";

const EMPTY_HIST = { cpu: [], gpu: [], mem: [], net: [], io: [] };

const fmtPct = (v) => `${Math.max(0, Math.round(v))}%`;
const fmtMem = (v) => `${Math.max(0, v).toFixed(2)}MB`;
const makeFmtNet = (netIsMB, netUnit) => (v) =>
  netIsMB ? `${Math.max(0, v).toFixed(2)}${netUnit}` : `${Math.max(0, Math.round(v))}${netUnit}`;

const ProcessCard = {
  name: "ProcessCard",
  components: { Sparkline, AnimatedNumber },
  props: {
    item: { type: Object, required: true },
  },
  computed: {
    store() {
      return store;
    },
    s() {
      return this.item;
    },
    status() {
      return statusInfo(this.item);
    },
    cardClasses() {
      const s = this.s;
      return {
        hung: !!s.hung,
        "row-disabled": !!s.disabled || s.status === "disabled",
        "status-running-bg": !s.hung && !s.disabled && s.status === "running",
        "status-started-bg": !s.hung && !s.disabled && s.status === "started",
        "status-stopped-bg": !s.hung && !s.disabled && (s.status === "stopped" || s.status === "unknown"),
        "status-disabled-bg": !!s.disabled || s.status === "disabled",
      };
    },
    hist() {
      return store.history[this.item.name] || EMPTY_HIST;
    },
    cpuVal() {
      return parseFloat(this.s.cpu || "0") || 0;
    },
    gpuVal() {
      return parseFloat(this.s.gpu || "0") || 0;
    },
    memVal() {
      return parseFloat(this.s.mem_mb || "0") || 0;
    },
    netVal() {
      return parseFloat(this.s.net_kbs || "0") || 0;
    },
    ioVal() {
      return parseFloat(this.s.io_kbs || "0") || 0;
    },
    netUnit() {
      return (store.snapshot?.net_unit || "KB").toUpperCase();
    },
    netIsMB() {
      return this.netUnit === "MB";
    },
    netFormat() {
      return makeFmtNet(this.netIsMB, this.netUnit);
    },
    pctFmt() {
      return fmtPct;
    },
    memFmt() {
      return fmtMem;
    },
    pidNum() {
      return Number(this.s.pid);
    },
    pidVisible() {
      return Number.isFinite(this.pidNum) && this.pidNum > 0;
    },
    pidFormat() {
      return (v) => `${Math.max(0, Math.round(v))}`;
    },
    startDisabled() {
      return !(this.s.status !== "running" && this.s.status !== "started");
    },
  },
  methods: {
    glowStyle(value, max) {
      return { opacity: Math.min(1, Math.abs(value) / max) };
    },
    metricTitle(kind, kv, mv) {
      return `${kind}: ${kv.toFixed(1)} KB/s | ${mv.toFixed(2)} MB/s`;
    },
    cpuTitle() {
      return `CPU: ${this.cpuVal.toFixed(1)}%`;
    },
    gpuTitle() {
      return `GPU: ${this.gpuVal.toFixed(1)}%`;
    },
    memTitle() {
      return `RAM: ${this.memVal.toFixed(2)} MB`;
    },
    netTitle() {
      const kv = this.netIsMB ? this.netVal * 1024 : this.netVal;
      const mv = this.netIsMB ? this.netVal : this.netVal / 1024;
      return this.metricTitle("NET", kv, mv);
    },
    ioTitle() {
      const kv = this.netIsMB ? this.ioVal * 1024 : this.ioVal;
      const mv = this.netIsMB ? this.ioVal : this.ioVal / 1024;
      return this.metricTitle("IO", kv, mv);
    },
    run(action) {
      store.processAction(this.item.name, action);
    },
    async onToggleDisabled() {
      await store.toggleDisabled(this.item);
    },
    openEditor() {
      store.openProcessEditor(this.item.name);
    },
  },
  template: `
    <article
      class="process-card"
      :class="cardClasses"
      :data-name="s.name"
      @click="openEditor"
    >
      <div class="process-card__bg">
        <div class="metric-glow metric-glow--cpu" :style="glowStyle(cpuVal, 100)"></div>
        <div class="metric-glow metric-glow--gpu" :style="glowStyle(gpuVal, 100)"></div>
        <div class="metric-glow metric-glow--ram" :style="glowStyle(memVal, 100)"></div>
        <div class="metric-glow metric-glow--net" :style="glowStyle(netVal, 250)"></div>
        <div class="metric-glow metric-glow--io" :style="glowStyle(ioVal, 250)"></div>
      </div>
      <div class="process-card__top">
        <div class="process-card__identity">
          <div class="process-name">{{ s.name || '' }}</div>
          <div class="process-meta">
            <span class="process-type">{{ (s.type || '').toUpperCase() }}</span>
            <span class="drawer-chip process-status" :class="status.cls">{{ status.icon }} {{ status.label }}</span>
          </div>
        </div>
        <div class="process-actions" @click.stop>
          <label class="action-toggle neon-toggle" title="Disabled">
            <input
              type="checkbox"
              class="action-switch"
              :checked="!!s.disabled"
              :disabled="store.toggleBusy[s.name]"
              @change="onToggleDisabled"
            />
          </label>
          <button class="neon-btn neon-btn--icon neon-btn--folder" title="Open folder" aria-label="Open folder" @click="run('open-folder')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6.5A2.5 2.5 0 0 1 5.5 4h4.2a2 2 0 0 1 1.5.7l1 1.3H18.5A2.5 2.5 0 0 1 21 8.5v7A2.5 2.5 0 0 1 18.5 18h-13A2.5 2.5 0 0 1 3 15.5v-9Z"/></svg>
          </button>
          <button class="neon-btn neon-btn--icon neon-btn--restart" title="Restart" aria-label="Restart" @click="run('restart')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5a7 7 0 1 1-6.1 10.4M7 5H3v4"/><path d="M3 9c1.4-3.4 4.7-5.8 8.5-5.8A8.5 8.5 0 1 1 5.2 18.5"/></svg>
          </button>
          <button class="neon-btn neon-btn--icon neon-btn--stop" title="Stop" aria-label="Stop" @click="run('stop')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="6.5" y="6.5" width="11" height="11" rx="2.2"/></svg>
          </button>
          <button class="neon-btn neon-btn--icon neon-btn--start" title="Start" aria-label="Start" :disabled="startDisabled" @click="run('start')">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5.5v13l11-6.5-11-6.5Z"/></svg>
          </button>
        </div>
      </div>
      <div class="process-card__main">
        <div class="process-stat"><span>PID</span>
          <template v-if="pidVisible">
            <AnimatedNumber :value="pidNum" :format="pidFormat" />
          </template>
          <strong v-else>-</strong>
        </div>
        <div class="process-stat"><span>Started</span><strong>{{ s.started_at || '-' }}</strong></div>
        <div class="process-stat"><span>Uptime</span><strong>{{ s.uptime || '-' }}</strong></div>
        <div class="process-stat full"><span>Target</span><strong>{{ s.target || '' }}</strong></div>
      </div>
      <div class="process-metrics">
        <div class="metric-chip">
          <span>CPU</span>
          <AnimatedNumber :value="cpuVal" :format="pctFmt" klass="metric-val" />
          <div class="spark-wrap" :title="cpuTitle()"><Sparkline :values="hist.cpu" color="#67e8f9" /></div>
        </div>
        <div class="metric-chip">
          <span>GPU</span>
          <AnimatedNumber :value="gpuVal" :format="pctFmt" klass="metric-val" />
          <div class="spark-wrap" :title="gpuTitle()"><Sparkline :values="hist.gpu" color="#fca5a5" /></div>
        </div>
        <div class="metric-chip">
          <span>RAM</span>
          <AnimatedNumber :value="memVal" :format="memFmt" klass="metric-val" />
          <div class="spark-wrap" :title="memTitle()"><Sparkline :values="hist.mem" color="#a7f3d0" /></div>
        </div>
        <div class="metric-chip">
          <span>NET</span>
          <AnimatedNumber :value="netVal" :format="netFormat" klass="metric-val" />
          <div class="spark-wrap" :title="netTitle()"><Sparkline :values="hist.net" color="#c4b5fd" /></div>
        </div>
        <div class="metric-chip">
          <span>IO</span>
          <AnimatedNumber :value="ioVal" :format="netFormat" klass="metric-val" />
          <div class="spark-wrap" :title="ioTitle()"><Sparkline :values="hist.io" color="#f9d46b" /></div>
        </div>
      </div>
    </article>
  `,
};

export { ProcessCard, fmtPct, fmtMem };
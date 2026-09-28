import { store } from "../store.js";
import { screenLabel } from "../util.js";

const computeLayout = (screens, availW, maxH) => {
  if (!screens.length) return null;
  let minX = Infinity;
  let minY = Infinity;
  let maxX = -Infinity;
  let maxY = -Infinity;
  for (const s of screens) {
    if (s.x < minX) minX = s.x;
    if (s.y < minY) minY = s.y;
    if (s.x + s.width > maxX) maxX = s.x + s.width;
    if (s.y + s.height > maxY) maxY = s.y + s.height;
  }
  const virtW = maxX - minX;
  const virtH = maxY - minY;
  if (virtW <= 0 || virtH <= 0) return null;
  const scale = Math.min(availW / virtW, maxH / virtH);
  return { minX, minY, scale, height: virtH * scale };
};

const MonitorPicker = {
  name: "MonitorPicker",
  props: {
    modelValue: { type: Number, default: 0 },
  },
  emits: ["update:modelValue"],
  computed: {
    layout() {
      return computeLayout(store.screens, 240, 70);
    },
    boxes() {
      const l = this.layout;
      if (!l) return [];
      return store.screens.map((s) => {
        const left = (s.x - l.minX) * l.scale;
        const top = (s.y - l.minY) * l.scale;
        const w = Math.max(s.width * l.scale - 4, 24);
        const h = Math.max(s.height * l.scale - 4, 16);
        return { screen: s, left, top, w, h };
      });
    },
    current() {
      if (this.modelValue > 0) return this.modelValue;
      const primary = store.screens.find((s) => s.primary);
      return primary ? primary.index : 0;
    },
    containerStyle() {
      return this.layout ? { height: `${Math.round(this.layout.height)}px` } : {};
    },
  },
  methods: {
    title(s) {
      return screenLabel(s);
    },
    pick(index) {
      this.$emit("update:modelValue", index);
    },
  },
  template: `
    <div class="monitor-picker" :style="containerStyle">
      <template v-if="boxes.length">
        <div
          v-for="b in boxes"
          :key="b.screen.index"
          class="monitor-pick"
          :class="{ selected: current === b.screen.index }"
          :style="{ left: b.left + 'px', top: b.top + 'px', width: b.w + 'px', height: b.h + 'px' }"
          :title="title(b.screen)"
          @click="pick(b.screen.index)"
        >
          <div class="monitor-pick-num">{{ b.screen.index }}</div>
          <div v-if="b.screen.primary" class="monitor-pick-badge">P</div>
        </div>
      </template>
      <span v-else>No screens</span>
    </div>
  `,
};

const ScreenViz = {
  name: "ScreenViz",
  data() {
    return { width: 320 };
  },
  computed: {
    layout() {
      return computeLayout(store.screens, Math.max(this.width - 32, 200), 260);
    },
    emptyText() {
      return store.screensError || "No screens detected";
    },
    boxes() {
      const l = this.layout;
      if (!l) return [];
      return store.screens.map((s) => {
        const left = (s.x - l.minX) * l.scale;
        const top = (s.y - l.minY) * l.scale;
        const w = Math.max(s.width * l.scale - 4, 50);
        const h = Math.max(s.height * l.scale - 4, 34);
        return { screen: s, left, top, w, h };
      });
    },
    containerStyle() {
      return this.layout ? { position: "relative", height: `${Math.round(this.layout.height)}px` } : {};
    },
  },
  mounted() {
    this.measure();
    window.addEventListener("resize", this.measure);
  },
  beforeUnmount() {
    window.removeEventListener("resize", this.measure);
  },
  methods: {
    measure() {
      this.width = this.$refs.root ? this.$refs.root.clientWidth : 320;
    },
    title(s) {
      return screenLabel(s);
    },
  },
  template: `
    <div ref="root" class="screen-list" :style="containerStyle">
      <template v-if="boxes.length">
        <div
          v-for="b in boxes"
          :key="b.screen.index"
          class="monitor-viz"
          :style="{ left: b.left + 'px', top: b.top + 'px', width: b.w + 'px', height: b.h + 'px' }"
          :title="title(b.screen)"
        >
          <div class="monitor-num">{{ b.screen.index }}</div>
          <div class="monitor-name">{{ b.screen.name || ('DISPLAY' + b.screen.index) }}</div>
          <div class="monitor-info">{{ b.screen.width }}x{{ b.screen.height }}</div>
          <div class="monitor-pos">x={{ b.screen.x }} y={{ b.screen.y }}</div>
          <div v-if="b.screen.primary" class="monitor-badge">PRIMARY</div>
        </div>
      </template>
      <span v-else>{{ emptyText }}</span>
    </div>
  `,
};

export { MonitorPicker, ScreenViz };
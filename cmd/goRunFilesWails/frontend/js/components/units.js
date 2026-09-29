import { clamp, toFiniteOr } from "../util.js";

let seq = 0;

const W = 90;
const H = 26;

const hexToRgb = (hex) => {
  const m = /^#?([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(String(hex).trim());
  if (!m) return null;
  const h = m[1].length === 3 ? m[1].split("").map((c) => c + c).join("") : m[1];
  const n = parseInt(h, 16);
  return `${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}`;
};

const withAlpha = (hex, a) => {
  const rgb = hexToRgb(hex);
  return rgb ? `rgba(${rgb}, ${a})` : String(hex);
};

const Sparkline = {
  name: "Sparkline",
  props: {
    values: { type: Array, default: () => [] },
    color: { type: String, default: "#67e8f9" },
  },
  data() {
    return { gradId: `grad-${++seq}` };
  },
  computed: {
    viewBox() {
      return `0 0 ${W} ${H}`;
    },
    points() {
      const v = this.values;
      if (!v.length) return `0,${H} ${W},${H}`;
      const n = Math.max(v.length, 2);
      const step = W / (n - 1);
      return v
        .map((x, i) => {
          const px = i * step;
          const py = H - (clamp(toFiniteOr(x, 0), 0, 100) / 100) * H;
          return `${px.toFixed(2)},${py.toFixed(2)}`;
        })
        .join(" ");
    },
    area() {
      return `0,${H} ${this.points} ${W},${H}`;
    },
    fill() {
      return `url(#${this.gradId})`;
    },
    glowStyle() {
      const c = this.color;
      return {
        filter:
          `drop-shadow(0 0 0.1rem ${withAlpha(c, 0.95)})` +
          ` drop-shadow(0 0 0.3rem ${withAlpha(c, 0.6)})` +
          ` drop-shadow(0 0 0.7rem ${withAlpha(c, 0.3)})`,
      };
    },
  },
  template: `
    <svg :viewBox="viewBox" width="100%" height="100%" preserveAspectRatio="none" class="spark" :style="glowStyle">
      <defs>
        <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :stop-color="color" stop-opacity="0.65" />
          <stop offset="45%" :stop-color="color" stop-opacity="0.26" />
          <stop offset="100%" :stop-color="color" stop-opacity="0.08" />
        </linearGradient>
      </defs>
      <polygon :points="area" :fill="fill" />
      <polyline
        :points="points"
        fill="none"
        :stroke="color"
        stroke-width="1.2"
        stroke-linecap="round"
        stroke-linejoin="round"
        vector-effect="non-scaling-stroke"
      />
    </svg>
  `,
};

const AnimatedNumber = {
  name: "AnimatedNumber",
  props: {
    value: { type: Number, default: 0 },
    format: {
      type: Function,
      default: (v) => String(Math.round(v)),
    },
    duration: { type: Number, default: 320 },
    klass: { type: String, default: "" },
  },
  data() {
    return { display: "", last: undefined };
  },
  watch: {
    value: {
      immediate: true,
      handler(to) {
        this.animate(to);
      },
    },
  },
  methods: {
    animate(to) {
      const from = typeof this.last === "number" && Number.isFinite(this.last) ? this.last : to;
      this.last = to;
      if (!Number.isFinite(to) || to === from || Math.abs(to - from) < 0.0001) {
        this.display = this.format(to);
        return;
      }
      const start = performance.now();
      const duration = this.duration;
      const step = (now) => {
        const p = Math.min(1, (now - start) / duration);
        const eased = 1 - Math.pow(1 - p, 3);
        this.display = this.format(from + (to - from) * eased);
        if (p < 1) requestAnimationFrame(step);
      };
      requestAnimationFrame(step);
    },
  },
  template: `<strong :class="klass">{{ display }}</strong>`,
};

export { Sparkline, AnimatedNumber };
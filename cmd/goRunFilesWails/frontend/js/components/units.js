import { clamp, toFiniteOr } from "../util.js";

let seq = 0;

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
      return "0 0 90 26";
    },
    points() {
      const w = 90;
      const h = 26;
      const v = this.values;
      const n = Math.max(v.length, 2);
      const step = w / (n - 1);
      return v
        .map((x, i) => {
          const px = i * step;
          const py = h - (clamp(toFiniteOr(x, 0), 0, 100) / 100) * h;
          return `${px.toFixed(2)},${py.toFixed(2)}`;
        })
        .join(" ");
    },
    area() {
      return `0,26 ${this.points} 90,26`;
    },
    fill() {
      return `url(#${this.gradId})`;
    },
  },
  template: `
    <svg :viewBox="viewBox" width="90" height="26" class="spark">
      <defs>
        <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :stop-color="color" stop-opacity="0.55" />
          <stop offset="100%" :stop-color="color" stop-opacity="0" />
        </linearGradient>
      </defs>
      <polygon :points="area" :fill="fill" />
      <polyline :points="points" fill="none" :stroke="color" stroke-width="2" />
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
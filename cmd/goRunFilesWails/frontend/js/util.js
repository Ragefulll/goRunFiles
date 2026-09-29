const clamp = (v, min, max) => Math.max(min, Math.min(max, v));

const toFiniteOr = (v, fallback) => {
  const n = Number(v);
  return Number.isFinite(n) ? n : fallback;
};

const STATUS_META = {
  running: { label: "Running", icon: "\ue768", cls: "status-running" },
  started: { label: "Starting", icon: "\ue895", cls: "status-started" },
  stopped: { label: "Stopped", icon: "\ue71a", cls: "status-stopped" },
  disabled: { label: "Disabled", icon: "\ue769", cls: "status-disabled" },
  hung: { label: "Hung", icon: "\ue7ba", cls: "status-hung" },
  unknown: { label: "Unknown", icon: "\ue946", cls: "status-unknown" },
};

const statusInfo = (it) => {
  if (!it) return STATUS_META.unknown;
  if (it.hung) return STATUS_META.hung;
  if (it.disabled || it.status === "disabled") return STATUS_META.disabled;
  return STATUS_META[it.status] || STATUS_META.unknown;
};

const screenLabel = (screen) => {
  if (!screen) return "";
  const parts = [
    `${screen.index}: ${screen.name || `Screen ${screen.index}`}`,
    `${screen.width}x${screen.height}`,
    `x=${screen.x}`,
    `y=${screen.y}`,
  ];
  if (screen.primary) parts.push("primary");
  return parts.join(" | ");
};

const clone = (v) => (v == null ? v : JSON.parse(JSON.stringify(v)));

export { clamp, toFiniteOr, STATUS_META, statusInfo, screenLabel, clone };
import { AppHeader, ProcessList } from "./js/components/ProcessList.js";
import { ProcessDrawer } from "./js/components/ProcessDrawer.js";
import { AuthModal, UpdateModal } from "./js/components/Modals.js";
import { ConfigModal } from "./js/components/ConfigModal.js";
import { store } from "./js/store.js";

const { createApp } = window.Vue;

let fontSizeHtml = "";
let fontAnimTimer = null;

const triggerFontSizeAnim = () => {
  const html = document.documentElement;
  html.classList.remove("font-size-anim");
  void html.offsetWidth;
  html.classList.add("font-size-anim");
  if (fontAnimTimer) clearTimeout(fontAnimTimer);
  fontAnimTimer = setTimeout(() => html.classList.remove("font-size-anim"), 260);
};

// Checkbox pulse animation (kept as a plain DOM hook).
document.addEventListener("change", (e) => {
  const el = e.target;
  if (el && el.type === "checkbox") {
    el.classList.remove("pulse");
    void el.offsetWidth;
    el.classList.add("pulse");
  }
});

const App = {
  name: "App",
  components: { AppHeader, ProcessList, ProcessDrawer, AuthModal, ConfigModal, UpdateModal },
  computed: {
    store() {
      return store;
    },
  },
  methods: {
    onResize() {
      const w = window.innerWidth;
      const h = window.innerHeight;
      store.windowLabel = `${w}x${h}`;
      const px = `${(w / 2050) * 11}px`;
      if (px !== fontSizeHtml) {
        fontSizeHtml = px;
        document.documentElement.style.fontSize = px;
        triggerFontSizeAnim();
      }
    },
  },
  mounted() {
    this.onResize();
    window.addEventListener("resize", this.onResize);
  },
  beforeUnmount() {
    window.removeEventListener("resize", this.onResize);
  },
  template: `
    <div class="wrap">
      <AppHeader />
      <ProcessList />
    </div>
    <ProcessDrawer />
    <AuthModal />
    <ConfigModal />
    <UpdateModal />
  `,
};

createApp(App).mount("#app");
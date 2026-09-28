import { store } from "../store.js";
import { ProcessCard } from "./ProcessCard.js";

const AppHeader = {
  name: "AppHeader",
  computed: {
    store() {
      return store;
    },
  },
  template: `
    <header class="head">
      <div class="block">
        <strong class="textSizeDebug">{{ store.windowLabel }}</strong>
        <span>&nbsp;</span>
      </div>
      <div class="block title">
        <strong>ART3D Process Monitor</strong>
        <span>&nbsp;</span>
      </div>
      <div class="block meta"><span>Version:</span><strong>{{ store.version }}</strong></div>
      <div class="block meta"><span>Net:</span><strong :title="store.netStatusTitle">{{ store.netStatus }}</strong></div>
      <div class="block meta"><span>Net Debug:</span><strong>{{ store.netDebug }}</strong></div>
      <div class="block meta"><span>Updated:</span><strong>{{ store.updated }}</strong></div>
    </header>
  `,
};

const ProcessList = {
  name: "ProcessList",
  components: { ProcessCard },
  computed: {
    store() {
      return store;
    },
    checkBtnLabel() {
      return store.checkProcessRunning ? "Stop Check Process" : "Start Check Process";
    },
    checkBtnTitle() {
      return store.checkProcessRunning ? "Stop process checks" : "Start process checks";
    },
  },
  watch: {
    "store.consoleOpened": {
      handler(open) {
        if (open && this.$refs.ta) {
          this.$nextTick(() => {
            this.$refs.ta.scrollTop = 0;
          });
        }
      },
    },
  },
  mounted() {
    store.start();
  },
  template: `
    <section class="panel">
      <div class="panel-head">
        <h2>Processes</h2>
        <div class="panel-actions">
          <button class="panel-actions__button" title="Restart all" aria-label="Restart all" @click="store.restartAll()">Restart All</button>
          <button class="panel-actions__button" title="Test auto-restart" aria-label="Test auto-restart" @click="store.restartAutoManual()">Auto Restart</button>
          <button class="panel-actions__button" title="Stop all" aria-label="Stop all" @click="store.stopAll()">Stop All</button>
          <button class="panel-actions__button" title="Kill all cmd.exe" aria-label="Kill all cmd.exe" @click="store.killCMD()">Kill CMD</button>
          <button
            class="panel-actions__button"
            :class="{ active: !store.checkProcessRunning }"
            :disabled="store.checkToggleBusy"
            :title="checkBtnTitle"
            :aria-label="checkBtnTitle"
            @click="store.toggleCheckProcess()"
          >{{ checkBtnLabel }}</button>
          <button class="panel-actions__button" title="Kill all node.exe" aria-label="Kill all node.exe" @click="store.killNode()">Kill Node</button>
          <button class="panel-actions__button" :class="{ active: store.consoleOpened }" title="Консоль ошибок" aria-label="Консоль ошибок" @click="store.toggleConsole()">📋</button>
          <button class="panel-actions__button" title="Lock config" aria-label="Lock config" @click="store.lockConfig()">Lock</button>
          <button class="panel-actions__button" title="Настройки" aria-label="Настройки" @click="store.toggleConfigOpen()">⚙</button>
        </div>
      </div>
      <div class="process-cards">
        <ProcessCard v-for="item in store.items" :key="item.name" :item="item" />
      </div>
      <div class="error-console" :class="{ 'is-open': store.consoleOpened }">
        <textarea ref="ta" readonly spellcheck="false" aria-label="Логи ошибок" :value="store.errorLog.join('\\n')"></textarea>
      </div>
    </section>
  `,
};

export { AppHeader, ProcessList };
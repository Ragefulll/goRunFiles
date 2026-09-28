import { store } from "../store.js";

const AuthModal = {
  name: "AuthModal",
  computed: {
    store() {
      return store;
    },
  },
  template: `
    <div class="modal" v-if="store.authModalOpen">
      <div class="modal-backdrop" @click="store.closeAuthModal()"></div>
      <div class="modal-card">
        <div class="modal-head">
          <div>Config Access</div>
          <button title="Закрыть" @click="store.closeAuthModal()">✕</button>
        </div>
        <div class="modal-body">
          <label>Password
            <input type="password" v-model="store.password" placeholder="Введите пароль" @keyup.enter="store.unlockConfig()" />
          </label>
          <p v-if="store.authError" class="auth-error" :style="{ color: '#f87171', marginTop: '0.6rem' }">{{ store.authError }}</p>
        </div>
        <div class="modal-actions">
          <button @click="store.unlockConfig()">Unlock</button>
          <button @click="store.closeAuthModal()">Cancel</button>
        </div>
      </div>
    </div>
  `,
};

const UpdateModal = {
  name: "UpdateModal",
  computed: {
    store() {
      return store;
    },
    bgStyle() {
      if (store.update.indeterminate) return {};
      return { width: `${store.update.pct || 0}%` };
    },
  },
  template: `
    <div class="update-pop" v-if="store.update.visible">
      <div class="update-card" :class="store.update.cls">
        <div class="update-bg" :class="{ indeterminate: store.update.indeterminate }" :style="bgStyle"></div>
        <div class="update-icon" v-html="store.update.icon"></div>
        <div class="update-body">
          <div class="update-head">
            <h2>{{ store.update.title }}</h2>
            <p>{{ store.update.msg }}</p>
          </div>
          <div class="modal-row"><span>{{ store.update.detail }}</span><strong>{{ store.update.pct }}</strong></div>
        </div>
      </div>
    </div>
  `,
};

export { AuthModal, UpdateModal };
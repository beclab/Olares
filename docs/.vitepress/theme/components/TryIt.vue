<template>
  <div class="tryit">
    <div class="tryit-head">
      <span class="tryit-badge">Try it</span>
      <code class="tryit-route">POST /api/{{ endpoint }}</code>
      <span v-if="endpoint === 'getPayment'" class="tryit-auth">client_secret — no API key</span>
      <span v-else class="tryit-auth">public — no API key</span>
    </div>

    <div v-if="endpoint === 'getPayment'" class="tryit-inputs">
      <input v-model="intentId" class="tryit-input" placeholder="intent_id (pi_…)" spellcheck="false" />
      <input v-model="clientSecret" class="tryit-input" placeholder="client_secret (pi_…_secret_…)" spellcheck="false" />
    </div>

    <div class="tryit-actions">
      <template v-if="sameOrigin">
        <button class="tryit-send" :disabled="loading || !ready" @click="send">
          {{ loading ? 'Sending…' : 'Send request' }}
        </button>
        <span class="tryit-target">{{ base }}/api/{{ endpoint }}</span>
      </template>
      <span v-else class="tryit-note">
        Live calls run on the production docs site (same origin as the gateway). This preview is on
        <code>{{ origin }}</code>, where the browser blocks the cross-origin call.
      </span>
    </div>

    <div v-if="error" class="tryit-resp tryit-error">{{ error }}</div>
    <pre v-else-if="responseText" class="tryit-resp"><code>{{ responseText }}</code></pre>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

const DEFAULT_BASE = 'https://www.olares.com/payment';

const props = withDefaults(defineProps<{ endpoint: 'ping' | 'getPayment'; base?: string }>(), {
  base: DEFAULT_BASE,
});

const intentId = ref('');
const clientSecret = ref('');
const loading = ref(false);
const responseText = ref('');
const error = ref('');

const ready = computed(() =>
  props.endpoint === 'ping' ? true : intentId.value.trim() !== '' && clientSecret.value.trim() !== '',
);

// Live calls are same-origin only: the production docs site and the gateway both
// live on www.olares.com. Any other origin (local preview, staging) is blocked by
// browser CORS — the gateway sends no CORS headers by design — so we degrade to a
// note instead of a broken button. SSR-safe: no location at setup time.
const origin = ref('');
// An explicit base override (local testing) bypasses the gate deliberately.
const sameOrigin = computed(() => props.base !== DEFAULT_BASE || origin.value === 'https://www.olares.com');
onMounted(() => {
  origin.value = location.origin;
});

async function send() {
  loading.value = true;
  responseText.value = '';
  error.value = '';
  const body =
    props.endpoint === 'ping'
      ? '{}'
      : JSON.stringify({ intent_id: intentId.value.trim(), client_secret: clientSecret.value.trim() });
  try {
    const res = await fetch(`${props.base}/api/${props.endpoint}`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body,
    });
    const text = await res.text();
    try {
      responseText.value = `HTTP ${res.status}\n${JSON.stringify(JSON.parse(text), null, 2)}`;
    } catch {
      responseText.value = `HTTP ${res.status}\n${text}`;
    }
  } catch (e) {
    error.value = `Request failed: ${(e as Error).message}`;
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.tryit {
  margin: 16px 0;
  padding: 16px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  background: var(--vp-c-bg-soft);
}
.tryit-head {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.tryit-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
}
.tryit-route {
  font-size: 13px;
  font-weight: 600;
}
.tryit-auth {
  font-size: 12px;
  color: var(--vp-c-text-2);
}
.tryit-inputs {
  display: flex;
  gap: 8px;
  margin-top: 12px;
  flex-wrap: wrap;
}
.tryit-input {
  flex: 1 1 220px;
  padding: 8px 10px;
  font-size: 13px;
  font-family: var(--vp-font-family-mono);
  border: 1px solid var(--vp-c-divider);
  border-radius: 6px;
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
}
.tryit-input:focus {
  outline: none;
  border-color: var(--vp-c-brand-1);
}
.tryit-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  flex-wrap: wrap;
}
.tryit-send {
  padding: 7px 16px;
  font-size: 13px;
  font-weight: 600;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  background: var(--vp-c-brand-1);
  color: var(--vp-c-white);
}
.tryit-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.tryit-target {
  font-size: 12px;
  color: var(--vp-c-text-3);
  font-family: var(--vp-font-family-mono);
  word-break: break-all;
}
.tryit-resp {
  margin-top: 12px;
  padding: 12px;
  border-radius: 6px;
  background: var(--vp-c-bg);
  border: 1px solid var(--vp-c-divider);
  font-size: 12.5px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
.tryit-error {
  color: var(--vp-c-danger-1);
  font-size: 13px;
}
.tryit-note {
  font-size: 12.5px;
  color: var(--vp-c-text-2);
}
.tryit-note code {
  font-size: 12px;
}
</style>

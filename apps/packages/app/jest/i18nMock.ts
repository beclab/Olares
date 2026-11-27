/**
 * The real boot file builds a vue-i18n instance and reads the browser locale at
 * import time, which a node suite has no way to satisfy. Suites that only need
 * a label to be *some* stable string get the key back.
 */
export const t = (key: string) => key;

export const i18n = { global: { t } };

export default t;

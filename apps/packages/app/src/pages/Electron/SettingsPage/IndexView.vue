<template>
	<div class="indexView full-height q-pl-md">
		<Settings v-if="settingMode == '1'" />
		<Account v-if="settingMode == '4'" />
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import Settings from './SettingsPage.vue';
import Account from './Account.vue';

export default defineComponent({
	name: 'IndexView',
	props: {
		settingMode: {
			type: String,
			required: true
		}
	},
	components: {
		Settings,
		Account
	},
	setup() {
		return {};
	}
});
</script>

<style lang="scss" scoped>
.indexView {
	width: 100%;
	background: $white;
	overflow: scroll;
}
</style>

<template>
	<q-icon
		v-if="description"
		size="16px"
		class="q-ml-sm"
		:class="classes"
		name="sym_r_help"
		@click.stop.prevent
	>
		<!-- Touch devices have no hover: tap to open, tap outside to dismiss. -->
		<q-menu
			v-if="isTouch"
			class="settings-tooltip-menu bg-tooltip-bg"
			max-width="280px"
			anchor="bottom middle"
			self="top left"
		>
			<span class="text-body3 text-ink-tooltip">{{ description }}</span>
		</q-menu>
		<q-tooltip v-else max-width="280px" anchor="bottom middle" self="top left">
			<span class="text-body3 text-ink-tooltip">{{ description }}</span>
		</q-tooltip>
	</q-icon>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';

defineProps({
	description: {
		type: String,
		required: false,
		default: ''
	},
	classes: {
		required: false,
		default: 'text-ink-3',
		type: String
	}
});

const $q = useQuasar();
const isTouch = $q.platform.has.touch;
</script>

<style lang="scss">
// The q-menu popup is teleported to <body>, so styles can't be scoped.
// Match Quasar's q-tooltip look (padding/radius, no elevation shadow).
.settings-tooltip-menu {
	padding: 4px 8px;
	border-radius: 4px;
	box-shadow: none;
}
</style>

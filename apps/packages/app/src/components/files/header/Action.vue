<template>
	<button @click="action" :aria-label="label" :title="label" class="action">
		<!-- <i class="material-icons">{{ icon }}</i> -->
		<q-icon :name="`sym_r_${icon}`" color="ink-2" size="18px"></q-icon>
		<span>{{ label }}</span>
		<span v-if="counter && counter > 0" class="counter">{{ counter }}</span>
	</button>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { useDataStore } from '../../../stores/data';

export default defineComponent({
	name: 'ActionComponent',
	props: {
		icon: {
			type: String,
			required: false
		},
		label: {
			type: String,
			required: false
		},
		counter: {
			type: Number,
			required: false
		},
		show: {
			type: String,
			required: false
		}
	},
	setup(props, context) {
		const store = useDataStore();
		const action = () => {
			if (props.show) {
				store.showHover(props.show);
			}

			context.emit('action');
		};

		return {
			action
		};
	}
});
</script>

<style lang="scss" scoped>
#previewer header .action:hover {
	background-color: $btn-bg-pressed;
}
</style>

<template>
	<path-location style="height: 60px">
		<template v-slot:start>
			<q-icon size="20px" name="sym_r_lock" />
			<div
				class="text-body2 secret-label"
				style="margin-left: 4px; max-width: calc(100% - 100px)"
			>
				{{ secret?.Key }}
			</div>
		</template>
		<template v-slot:midden>
			<div class="text-body2 secret-label">????</div>
		</template>
		<template v-slot:end>
			<div class="row">
				<q-icon
					class="cursor-pointer"
					style="margin-right: 12px"
					size="20px"
					name="sym_r_edit_square"
					@click="emit('onUpdate', secret?.Key)"
				/>
				<q-icon
					class="cursor-pointer"
					size="20px"
					name="sym_r_delete"
					@click="emit('onDelete', secret?.Key)"
				/>
			</div>
		</template>
	</path-location>
</template>

<script lang="ts" setup>
import PathLocation from '../PathLocation.vue';
import { Secret } from 'src/constant/index';
import { PropType } from 'vue';

defineProps({
	secret: {
		type: Object as PropType<Secret>,
		require: true
	},
	appName: {
		type: String,
		require: true
	}
});

const emit = defineEmits(['onUpdate', 'onDelete']);
</script>

<style scoped lang="scss">
.secret-label {
	color: $ink-1;
}
</style>

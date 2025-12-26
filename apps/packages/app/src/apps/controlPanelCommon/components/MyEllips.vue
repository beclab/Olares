<template>
	<text-clamp class="my-ellips-container" :text="text" :max-lines="2">
		<template #before="{ toggle, clamped, expanded }">
			<QButtonStyle size="md">
				<q-btn
					dense
					unelevated
					no-caps
					class="my-ellips-icon text-grey-10"
					v-if="!(!clamped && !expanded)"
					:icon="expanded ? 'arrow_drop_down' : 'arrow_right'"
					@click="toggle"
				/>
			</QButtonStyle>
		</template>
	</text-clamp>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import TextClamp from 'vue3-text-clamp';
import QButtonStyle from './QButtonStyle.vue';
interface Props {
	text: string;
}
withDefaults(defineProps<Props>(), {
	text: ''
});
</script>

<style lang="scss" scoped>
.my-ellips-container {
	&.text-clamp {
		position: relative;
		padding-left: 16px;
	}
	.my-ellips-icon {
		position: absolute;
		left: 0;
		top: 1px;
	}
}
</style>

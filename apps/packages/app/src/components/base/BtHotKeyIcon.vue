<template>
	<div
		class="row justify-center items-center text-ink-3 no-wrap"
		style="gap: 4px"
		v-if="keys.length > 0"
	>
		<q-icon v-if="showBoard && single" size="16px" name="sym_r_keyboard_alt" />
		<template v-for="item in keys" :key="item">
			<q-icon v-if="isIOS && item === 'shift'" size="16px" name="sym_r_shift" />
			<q-icon
				v-else-if="isIOS && (item === 'alt' || item === 'option')"
				size="16px"
				name="sym_r_keyboard_option_key"
			/>
			<q-icon
				v-else-if="isIOS && item === 'option'"
				size="16px"
				name="sym_r_keyboard_command_key"
			/>
			<q-icon
				v-else-if="isIOS && item === 'control'"
				size="16px"
				name="sym_r_keyboard_control_key"
			/>
			<q-icon
				v-else-if="isIOS && item === 'backspace'"
				size="16px"
				name="sym_r_backspace"
			/>
			<q-icon
				v-else-if="isIOS && item === 'tab'"
				size="16px"
				name="sym_r_keyboard_tab"
			/>
			<q-icon
				v-else-if="isIOS && item === 'enter'"
				size="16px"
				name="sym_r_keyboard_return"
			/>
			<q-icon
				v-else-if="
					isIOS && (item === 'capslock' || item === 'Caps Lock' || item === '⇪')
				"
				size="16px"
				name="sym_r_keyboard_capslock"
			/>
			<q-icon
				v-else-if="isIOS && item === 'space'"
				size="16px"
				name="sym_r_space_bar"
			/>
			<q-icon
				v-else-if="isIOS && item === 'up'"
				size="16px"
				name="sym_r_arrow_drop_up"
			/>
			<q-icon
				v-else-if="isIOS && item === 'down'"
				size="16px"
				name="sym_r_arrow_drop_down"
			/>
			<q-icon
				v-else-if="isIOS && item === 'left'"
				size="16px"
				name="sym_r_arrow_left"
			/>
			<q-icon
				v-else-if="isIOS && item === 'right'"
				size="16px"
				name="sym_r_arrow_right"
			/>

			<div class="text-capitalize text-body3" v-else>{{ item }}</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useQuasar } from 'quasar';

const props = defineProps({
	hotkey: {
		type: String,
		default: ''
	},
	showBoard: {
		type: Boolean,
		default: true
	}
});

const keys = computed(() => {
	if (props.hotkey.length > 0) {
		if (props.hotkey.includes('+')) {
			return props.hotkey.split('+');
		} else {
			return [props.hotkey];
		}
	} else {
		return [];
	}
});

const single = computed(() => {
	return keys.value.length === 1 && keys.value[0].length === 1;
});

const $q = useQuasar();

const isIOS = computed(() => {
	return (
		$q.platform.is.ios ||
		$q.platform.is.ipad ||
		$q.platform.is.mac ||
		$q.platform.is.safari
	);
});
</script>

<style scoped lang="scss"></style>

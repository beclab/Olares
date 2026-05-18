<template>
	<q-toggle
		v-if="props.showThemeToggle"
		v-model="isDark"
		@update:model-value="change"
	/>
</template>

<script setup lang="ts">
import { ref, defineProps, onMounted, watch } from 'vue';
import { useQuasar } from 'quasar';
import { themeModeName, ThemeDefinedMode } from '../types';

interface Props {
	showThemeToggle?: boolean;
	followSystem?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	showThemeToggle: false,
	followSystem: true
});

const $q = useQuasar();
const isDark = ref(false);

if (props.showThemeToggle) {
	$q.dark.set(true);
	isDark.value = true;
}

const change = () => {
	$q.dark.toggle();
};

function getCookie(name: string) {
	if (typeof document !== 'undefined') {
		var strcookie = document.cookie;
		var arrcookie = strcookie.split('; ');
		for (var i = 0; i < arrcookie.length; i++) {
			var arr = arrcookie[i].split('=');
			if (arr[0] == name) {
				return arr[1];
			}
		}
	}

	return '';
}

function setTheme() {
	let themeValue = '';
	if ($q.cookies) {
		themeValue = $q.cookies.get(themeModeName);
	} else {
		themeValue = getCookie(themeModeName);
	}

	if (themeValue) {
		const theme = Number(themeValue);
		if (theme == ThemeDefinedMode.AUTO) {
			$q.dark.set('auto');
		} else {
			$q.dark.set(theme == ThemeDefinedMode.DARK);
		}
	}
}

const message = (event: any) => {
	if (props.followSystem && event.data.message === 'theme_apps_update') {
		if (event.data.info.theme.toString() === '1') {
			$q.dark.set(false);
		} else {
			$q.dark.set(true);
		}
	}
};

setTheme();

onMounted(() => {
	window && window.addEventListener('message', message);
});

watch(
	() => props.followSystem,
	(newValue) => {
		if (newValue) {
			setTheme();
		}
	}
);
</script>
<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
	name: 'BtTheme'
});
</script>
<style></style>

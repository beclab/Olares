import { useQuasar, colors, Dark } from 'quasar';
import { ref, watchEffect } from 'vue';

const { getPaletteColor } = colors;

function useColor(variable: string) {
	const $q = useQuasar();
	const dark = $q?.dark || Dark;
	const color = ref();
	try {
		color.value = getPaletteColor(variable);

		watchEffect(() => {
			color.value = dark.isActive
				? getPaletteColor(variable)
				: getPaletteColor(variable);
		});
		return { color };
	} catch (error) {
		return { color };
	}
}

export default useColor;

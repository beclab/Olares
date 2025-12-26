import { ref, reactive } from 'vue';

export const size = reactive({
	widths: [0]
});

export const initSize = () => {
	size.widths = [0];
};

export const updateWidths = (index: number, value: number) => {
	size.widths[index] = value;
};

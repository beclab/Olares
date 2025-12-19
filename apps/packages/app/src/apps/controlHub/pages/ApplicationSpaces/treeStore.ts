import { get } from 'lodash';
import { computed, ref } from 'vue';

export const myTreeRef = ref();

export const selectedNodes = computed(
	() => get(myTreeRef.value, 'selectedNodes'),
	{}
);

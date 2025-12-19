import { ref } from 'vue';

export const encrypting = ref(true);

export const show = () => {
	encrypting.value = false;
};

export const hide = () => {
	encrypting.value = true;
};

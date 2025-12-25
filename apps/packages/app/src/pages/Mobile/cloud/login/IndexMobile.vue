<template>
	<LoginForm @login-success="loginSuccess" />
</template>

<script lang="ts" setup>
import LoginForm from './LoginForm.vue';
import { useRouter } from 'vue-router';
const router = useRouter();

const loginSuccess = () => {
	emit('loginSuccess');
	router.go(-1);
};

const emit = defineEmits(['loginSuccess']);
</script>

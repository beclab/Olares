<template>
	<div class="add-file" @click="addFile">
		<img src="./../../../assets/add_file.svg" alt="add_file" />
	</div>
</template>

<script lang="ts" setup>
const emits = defineEmits(['addFile']);

const addFile = () => {
	emits('addFile');
};
</script>

<style lang="scss" scoped>
.add-file {
	position: fixed;
	right: 32px;
	// bottom: 64px;
	bottom: calc(env(safe-area-inset-bottom) + 46px + 64px);
	width: 44px;
	height: 44px;
	background-color: #ffeb3b;
	border-radius: 22px;
	display: flex;
	align-items: center;
	justify-content: center;
	// box-shadow: 0px 0px 24px 0 rgba($color: #51aeff, $alpha: 0.8);
	img {
		width: 16px;
		height: 16px;
	}
}
</style>

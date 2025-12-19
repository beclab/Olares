<template>
	<div class="connect-layout row justify-center items-center">
		<div class="text-caption connect-title">{{ $t('not_bound') }}</div>
	</div>
</template>

<style scoped lang="scss">
.connect-layout {
	padding: 0px 8px;
	border-radius: 4px;
	width: auto;
	height: 20px;
	margin-left: 4px;
	background-color: $red-1;

	.connect-title {
		color: $red;
	}
}
</style>

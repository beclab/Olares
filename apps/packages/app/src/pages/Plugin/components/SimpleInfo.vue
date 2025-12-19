<template>
	<q-card class="simple-card-container q-pa-sm" flat bordered>
		<div class="text-ody3 text-grey-10 text-wrapper ellipsis-2-lines">
			{{ data.text }}test
		</div>
		<div class="row items-center q-mt-sm">
			<div
				:class="[data.bordered ? 'bordered-wrapper' : '']"
				class="row items-center justify-center"
			>
				<q-img
					:src="data.img"
					:ratio="1"
					width="16px"
					spinner-size="xs"
					fit="contain"
					class="img"
				/>
			</div>

			<span class="text-body3 text-ink-2 q-ml-xs">{{ data.title }}</span>
		</div>
	</q-card>
</template>

<script setup lang="ts">
interface Props {
	data: {
		img: string;
		text: string;
		title: string;
		bordered?: boolean;
	};
}

withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.simple-card-container {
	border-color: $grey-2;
	border-radius: 8px;
	.bordered-wrapper {
		border-radius: 50%;
		padding: 2px;
		width: 16px;
		height: 16px;
		border: 1px solid $separator;
	}
	.img {
		border-radius: 50%;
	}
	.text-wrapper {
		height: 32px;
		line-height: 16px;
	}
}
</style>

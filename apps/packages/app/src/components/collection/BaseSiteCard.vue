<template>
	<div
		class="row items-center justify-between site-info-card no-wrap flex-gap-x-lg q-pa-md"
	>
		<div
			class="row items-center no-wrap flex-gap-x-md ellipsis"
			style="flex: 1"
		>
			<div class="img-wrapper row items-center">
				<q-img
					:src="data.icon || mask_group"
					:error-src="mask_group"
					width="40px"
					height="40px"
					crossorigin="anonymous"
					referrerpolicy="no-referrer"
					fit="cover"
				>
					<template #loading>
						<q-skeleton
							type="rect"
							square
							width="40px"
							height="40px"
							animation="fade"
						/>
					</template>
				</q-img>
			</div>
			<div class="ellipsis" style="">
				<div class="text-subtitle3 text-ink-1 ellipsis">
					{{ data.title }}
				</div>
				<div class="row no-wrap items-center flex-gap-xs q-mt-xs text-body3">
					<slot>
						<q-icon name="sym_r_link" size="16px" color="ink-6" />
						<div class="ellipsis">
							<span
								class="link-wrapper text-body3 text-ink-3"
								:href="data.url"
								target="_blank"
								>{{ data.url }}</span
							>
						</div>
					</slot>
				</div>
			</div>
		</div>
		<div style="width: 32px">
			<slot name="action"></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import { BaseSiteCardProps } from './collect';
import mask_group from 'src/assets/common/mask_group.svg';

withDefaults(defineProps<BaseSiteCardProps>(), {
	data: () => ({
		title: '',
		icon: '',
		url: ''
	})
});
</script>

<style lang="scss" scoped>
.site-info-card {
	border-radius: 12px;
	border: 1px solid $separator-2;
	.img-wrapper {
		flex: 0 0 40px;
		width: 40px;
		height: 40px;
		border-radius: 6px;
		overflow: hidden;
	}
	.link-wrapper {
		&:link {
			color: $ink-3;
		}
	}
}
</style>

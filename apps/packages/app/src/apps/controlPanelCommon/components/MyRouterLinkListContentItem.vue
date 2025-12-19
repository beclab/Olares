<template>
	<q-item
		class="my-menu-link q-py-none q-px-sm rounded-borders"
		active-class="my-menu-link-active"
		v-bind="$attrs"
	>
		<q-item-section
			v-if="data.img"
			avatar
			style="min-width: 0; padding-right: 8px"
		>
			<div class="avatar-wrapper">
				<div class="relative-position">
					<q-img :src="data.img" width="20px" height="20px" fit="cover" />
					<MyBadge
						v-if="data.status"
						:type="data.status"
						size="sm"
						class="my-badge-wrapper"
					></MyBadge>
				</div>
			</div>
		</q-item-section>

		<q-item-section>
			<q-item-label lines="1">
				<div class="text-body2 ellipsis">
					<span :class="{ active: data.id === active }">
						{{ data.title }}
					</span>
				</div>
			</q-item-label>
			<q-item-label caption lines="1">
				<div class="text-overline text-ink-3">
					{{ data.subTitle }}
				</div>
			</q-item-label>
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import MyBadge from './MyBadge.vue';

interface Props {
	data: any;
}

defineProps<Props>();
</script>

<style lang="scss" scoped>
.my-menu-link {
	color: $ink-1;
	min-height: 40px;
	::v-deep(.q-item__label + .q-item__label) {
		margin-top: 0;
	}
}
.my-menu-link-active {
	background-color: $blue-soft;
	color: $blue-default;
	.active {
		color: $blue-default;
	}
}
.avatar-wrapper {
	border: 1px solid rgba(0, 0, 0, 0);
	position: relative;
	.my-badge-wrapper {
		position: absolute;
		bottom: 0;
		right: 0;
	}
}
</style>

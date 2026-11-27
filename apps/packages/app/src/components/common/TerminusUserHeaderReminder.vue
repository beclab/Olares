<template>
	<div
		v-if="
			termipassStore.totalStatus?.isError == UserStatusActive.error ||
			$slots.errorcontent
		"
		class="error-content text-body3 row items-center justify-center q-py-sm"
	>
		<q-icon name="sym_r_error" size="20px" class="q-mr-sm" />
		<div v-if="!$slots.errorcontent" style="max-width: calc(100% - 70px)">
			<div v-if="termipassStore.totalStatus?.description">
				<div
					v-if="termipassStore.totalStatus.descriptionExI18nKey"
					class="reminder-action cursor-pointer"
					role="button"
					:aria-label="termipassStore.totalStatus.description"
					@click="itemClick"
				>
					{{ termipassStore.totalStatus.description }}
					<q-icon name="sym_r_chevron_right" size="16px" />
				</div>
				<div v-else>
					{{ termipassStore.totalStatus.description }}
				</div>
			</div>
		</div>
		<div v-else style="max-width: calc(100% - 70px)">
			<slot name="errorcontent" />
		</div>
	</div>
</template>

<script setup lang="ts">
import { UserStatusActive } from '../../utils/checkTerminusState';
import { getAppPlatform } from 'src/application/platform';
import { useTermipassStore } from '../../stores/termipass';
const termipassStore = useTermipassStore();

const itemClick = () => {
	const platform = getAppPlatform() as ReturnType<typeof getAppPlatform> & {
		userStatusUpdateAction?(): void;
	};
	platform.userStatusUpdateAction?.();
};
</script>

<style scoped lang="scss">
.error-content {
	background: $red-alpha;
	width: 100%;
	color: $red;
}

.reminder-action {
	text-decoration: underline;

	.q-icon {
		vertical-align: text-bottom;
		text-decoration: none;
	}
}
</style>

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
				<div v-if="termipassStore.totalStatus.descriptionEx">
					{{
						termipassStore.totalStatus.description.split(
							termipassStore.totalStatus.descriptionEx
						)[0]
					}}
					<span class="jump-subline-item" @click="itemClick">
						{{ termipassStore.totalStatus.descriptionEx }}
					</span>
					{{
						termipassStore.totalStatus.description.split(
							termipassStore.totalStatus.descriptionEx
						).length > 1
							? termipassStore.totalStatus.description.split(
									termipassStore.totalStatus.descriptionEx
							  )[1]
							: ''
					}}
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
import { getPlatform } from '@didvault/sdk/src/core';
import { TerminusCommonPlatform } from '../../platform/terminusCommon/terminalCommonPlatform';
import { useTermipassStore } from '../../stores/termipass';
const termipassStore = useTermipassStore();

const itemClick = () => {
	const platform = getPlatform() as unknown as TerminusCommonPlatform;
	console.log('platform ===>', platform);

	platform.userStatusUpdateAction();
};
</script>

<style scoped lang="scss">
.error-content {
	background: $red-alpha;
	width: 100%;
	color: $red;
}
</style>

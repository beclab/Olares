<template>
	<q-card-section
		class="row items-center justify-start q-pa-none header q-pl-sm"
	>
		<q-icon :name="showItemIcon(props.auditedItems.item.icon)" size="24px" />
		<div class="item-name q-ml-sm">
			{{
				props.auditedItems.item.name
					? props.auditedItems.item.name
					: t('new_item')
			}}
		</div>
	</q-card-section>

	<q-scroll-area
		ref="vaultItemRef"
		:thumb-style="thumbStyle as any"
		:visible="true"
		style="height: 44px; width: 100%; padding-left: 40px"
	>
		<q-card-section horizontal>
			<div
				v-for="(filed, index2) in props.auditedItems.item.fields"
				class="item-unit q-mr-sm"
				:key="`f` + index2"
			>
				<div class="text-blue-4" style="width: calc(100% + 20px)">
					<q-icon class="q-mr-sm" :name="filed.icon" />
					<span>
						{{ filed.name }}
					</span>
				</div>
				<div v-if="filed.value" class="text-color-title">
					<span v-if="filed.type === 'totp'">
						<Totp :secret="filed.value" />
					</span>
					<span v-else>
						{{ filed.format(true) }}
					</span>
				</div>
				<div v-else class="text-color-title">[empty]</div>
			</div>
		</q-card-section>
	</q-scroll-area>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { showItemIcon } from '../../../utils/utils';
import Totp from '../../Items/totp.vue';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'ReportCard',
	components: {
		Totp
	},
	props: {
		auditedItems: {
			type: Object,
			required: true
		}
	},

	setup(props) {
		const thumbStyle = ref({
			height: 0
		});

		const { t } = useI18n();

		return {
			thumbStyle,
			props,
			showItemIcon,
			t
		};
	}
});
</script>

<style scoped lang="scss">
.header {
	height: 28px;
}
.item-unit {
	div {
		white-space: nowrap;
	}
}
</style>

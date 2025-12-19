<template>
	<div class="card_wrap" :class="!selectIds ? 'select-not' : ''">
		<q-card
			clickable
			v-ripple
			@click="selectItem(vaultItem as ListItem)"
			:active="isSelected(vaultItem as ListItem)"
			active-class="text-blue"
			flat
			bordered
			class="vaultCard col-6 q-mx-md"
			:class="isSelected(vaultItem as ListItem) ? 'vaultCardActive' : ''"
			@mouseenter="handleMouseEnter(vaultItem.item.id)"
			@mouseleave="handleMouseLeave"
		>
			<q-scroll-area
				ref="vaultItemRef"
				:thumb-style="{ height: '0px' }"
				:visible="true"
				style="height: 65px; width: 100%"
			>
				<q-card-section horizontal>
					<div
						v-for="(filed, index2) in vaultItem.item.fields"
						class="item-unit cursor-pointer q-py-xs"
						:key="`f` + index2"
					>
						<div
							v-if="filed.value"
							class="text-ink-2 text-left item-unit-content q-ml-xs"
						>
							<span v-if="filed.type === 'totp'">
								<Totp2 :secret="filed.value" ref="myTotps" />
							</span>
							<span v-else>
								{{ filed.format(true) }}
							</span>
						</div>
						<div v-else class="text-body3">[{{ t('empty') }}]</div>
					</div>
				</q-card-section>
				<div class="ink-1 text-body-3 q-ml-lg q-pl-sm">
					{{ vaultItem.item.name }}
				</div>
			</q-scroll-area>
		</q-card>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, PropType } from 'vue';
import { useRoute } from 'vue-router';
import { VaultType } from '@didvault/sdk/src/core';
import { ListItem } from '@didvault/sdk/src/types';
import Totp2 from './totp2.vue';
import { notifyWarning } from '../../utils/notifyRedefinedUtil';

import { useI18n } from 'vue-i18n';

defineProps({
	vaultItem: {
		type: Object as PropType<ListItem>,
		required: true
	},
	selectIds: {
		type: Array as PropType<string[]>,
		required: false
	}
});

const emits = defineEmits(['selectItem']);

const Route = useRoute();
const isHoveredId = ref();
const myTotps = ref();

const { t } = useI18n();

async function selectItem(item: ListItem) {
	if (item.item.type === VaultType.TerminusTotp) {
		notifyWarning(t('vault_t.verification_message'));
		return false;
	}
	emits('selectItem', item);
}

function isSelected(item: ListItem): boolean {
	if (item && item.item.id === Route.params.itemid) {
		return true;
	}
	return false;
}

onMounted(async () => {});
onUnmounted(() => {});

const handleMouseEnter = (id: string) => {
	isHoveredId.value = id;
};

const handleMouseLeave = () => {
	isHoveredId.value = null;
};
</script>

<style lang="scss" scoped>
.item-unit {
	border-radius: 5px;
	margin-right: 4px;
	white-space: nowrap;
	position: relative;
	max-width: 180px;
	min-width: 70px;
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;

	.item-header {
		width: 100%;
		margin-bottom: 4px;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.item-unit-content {
		line-height: 1 !important;
		white-space: nowrap;
		span {
			display: inline-block;
			width: 100%;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
			font-size: 14px;
		}
	}

	.hideCopied {
		position: absolute;
		width: 100%;
		height: 100%;
		left: 0;
		top: 0;
		opacity: 0;
		border-radius: 8px;

		display: flex;
		align-items: center;
		justify-content: center;
		background: $grey-2;
		border: 1px solid $separator;
	}

	.copied {
		opacity: 1;
		border: 1px solid $yellow-default;
		background: $background-1;

		&:after {
			width: 100%;
			height: 100%;
			background: $yellow-alpha;
			content: '';
			position: absolute;
			top: 0;
			left: 0;
			z-index: 1;
		}
	}
}

.card_wrap {
	width: 100%;
	// border-bottom: 1px solid $separator;
	padding: 8px 0;
	position: relative;
	&::before {
		content: '';
		position: absolute;
		bottom: 0;
		right: 0;
		width: 100%;
		height: 1px;
		background-color: $separator;
		transition: width 0.5s ease;
	}

	&.select-not::before {
		width: 100%;
		width: calc(100% - 36px);
	}
}

.vaultCard {
	margin: 0px;
	border: 0;
	box-sizing: border-box;
	position: relative;
	padding: 8px;

	.tag-wrap {
		display: flex;
		align-items: center;
		justify-content: flex-end;

		.tag {
			border: 1px solid $ink-2;
			padding: 0 4px;
			border-radius: 4px;
			float: right;
			height: 20px;
			line-height: 20px;
			display: flex;
			align-items: center;
			justify-content: flex-start;

			.tag-name {
				max-width: 60px;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
			}
		}
	}

	.field-name {
		display: flex;
		align-items: center;
		justify-content: flex-start;

		.item-name {
			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;

			.label {
				overflow: hidden;
				white-space: nowrap;
				text-overflow: ellipsis;
			}

			.name {
				overflow: hidden;
				white-space: nowrap;
				text-overflow: ellipsis;
			}
		}
	}

	.west {
		position: absolute;
		left: 10px;
		bottom: 18px;
		margin: auto;
		z-index: 1;
		color: $ink-2;
		cursor: pointer;

		&:hover {
			color: $ink-1;
		}
	}

	.east {
		border-radius: 14px;
		position: absolute;
		right: 10px;
		bottom: 18px;
		margin: auto;
		z-index: 1;
		color: $ink-2;
		cursor: pointer;

		&:hover {
			color: $ink-1;
		}
	}
}
</style>

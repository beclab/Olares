<template>
	<div class="filter-item-root">
		<select-header :title="title" @reset="reset" />
		<div
			class="row items-center justify-between item text-body2 text-ink-1"
			:class="{
				'items-border': true
			}"
		>
			<div class="text-ink-2 text-body3 option-title">
				{{ !!optionTitle ? optionTitle : cOptionTitle }}
			</div>
			<q-icon
				:name="
					menuShow ? 'sym_r_keyboard_arrow_up' : 'sym_r_keyboard_arrow_down'
				"
				size="20px"
			/>
			<q-menu
				class="bg-background-2 q-pa-sm"
				fit
				:offset="offset"
				v-model="menuShow"
			>
				<q-list>
					<template v-for="option in options" :key="option.value">
						<q-item
							class="row items-center justify-start text-ink-3 q-pa-none q-px-sm"
							clickable
							dense
							style="height: 32px"
							@click.stop="itemClick(option)"
						>
							<terminus-check-box
								v-model="option.selected"
								:hookSelect="true"
								:activeImage="
									option.isAll && option.selected && hasNoSelected
										? 'img/checkbox/check_box_part.svg'
										: undefined
								"
								:label="option.label"
								:titleClasses="'text-body3'"
								@itemClick="itemClick(option)"
							/>
						</q-item>
						<q-separator v-if="option.isAll" />
					</template>
				</q-list>
			</q-menu>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, PropType, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import TerminusCheckBox from '../../common/TerminusCheckBox.vue';
import SelectHeader from './SelectHeader.vue';

interface MutipleItem {
	value: string | number;
	label: string;
	selected: boolean;
	isAll: boolean;
	isDefault: boolean;
}

const props = defineProps({
	options: {
		type: Object as PropType<MutipleItem[]>,
		require: true,
		default: [] as MutipleItem[]
	},
	title: {
		type: String,
		required: false,
		default: ''
	},
	optionTitle: {
		type: String,
		required: false,
		default: ''
	},
	offset: {
		type: Array,
		default: () => [0, 0],
		required: false
	}
});

const { t } = useI18n();

const cOptionTitle = computed(() => {
	const hasNoSelect = props.options.find((e) => e.selected == false);
	if (!hasNoSelect) {
		return t('my.all');
	}
	return props.options
		.filter((e) => !e.isAll && e.selected)
		.map((e) => e.label)
		.join(',');
});

const itemClick = (item: MutipleItem) => {
	if (item.isAll) {
		props.options.forEach((e) => {
			if (item.selected) {
				if (e.isAll || !e.selected) {
					return;
				}
				e.selected = false;
			} else {
				if (e.isAll || e.selected) {
					return;
				}
				e.selected = true;
			}
		});
		item.selected = !item.selected;
		return;
	} else {
		item.selected = !item.selected;
		if (isAllItem.value) {
			isAllItem.value.selected = hasSelectd.value;
		}
	}
};

const reset = () => {
	props.options.forEach((e) => {
		e.selected = true;
	});
};

const menuShow = ref(false);

const isAllItem = computed(() => {
	return props.options.find((e) => e.isAll);
});

const hasSelectd = computed(() => {
	return props.options.find((e) => !e.isAll && e.selected) != undefined;
});

const hasNoSelected = computed(() => {
	return props.options.find((e) => !e.isAll && !e.selected) != undefined;
});
</script>

<style scoped lang="scss">
.filter-item-root {
	height: 60px;
}

.item {
	margin-top: 4px;
	height: 32px;
	padding-left: 10px;
	padding-right: 10px;
	border-radius: 8px;
	background: $background-1;
	&:hover {
		background: $background-3;
	}

	.option-title {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		width: calc(100% - 20px);
	}
}

.items-border {
	border: solid 1px $separator;
}
</style>

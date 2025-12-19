<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('vault_t.new_vault_item')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:size="isWeb ? 'medium' : 'small'"
		:platform="isWeb ? 'web' : 'mobile'"
		@onSubmit="onOKClick"
	>
		<div class="s-dialog-plugin-web" v-if="isWeb">
			<q-card-section class="q-pt-xs q-px-none">
				<div class="text-left text-subtitle3 q-mt-md q-mb-sm text-ink-3">
					{{ t('vault_t.select_vault') }}
				</div>
				<div class="row align-center justify-center">
					<q-select
						class="select-vault"
						popup-content-class="options_selected_Account"
						standout
						dense
						emit-value
						map-options
						popup-no-route-dismiss
						v-model="model"
						:menu-offset="[0, 4]"
						dropdown-icon="sym_r_expand_more"
						:options="
							vaults.map((v) => ({
								value: v.id,
								label: `${v.org?.name || ''}${v.org?.name ? ' / ' : ''}${
									v.name
								}`
							}))
						"
						@update:model-value="changeModel"
						style="width: 100%"
					>
						<template
							v-slot:option="{ itemProps, opt, selected, toggleOption }"
						>
							<q-item v-bind="itemProps">
								<q-item-section>
									<q-item-label>{{ opt.label }}</q-item-label>
								</q-item-section>
								<q-item-section side>
									<q-checkbox
										:model-value="selected"
										checked-icon="sym_r_check_circle"
										unchecked-icon=""
										indeterminate-icon="help"
										@update:model-value="toggleOption(opt.label)"
									/>
								</q-item-section>
							</q-item>
						</template>
					</q-select>
				</div>
			</q-card-section>

			<div class="text-subtitle3 text-ink-3 text-left q-mb-sm q-px-none">
				{{ t('vault_t.what_kind_of_item_you_would_like_to_add') }}
			</div>
			<q-card-section
				class="row q-col-gutter-md q-px-none"
				style="padding-top: 0"
			>
				<div v-for="(item, index) in templates" class="col-6" :key="index">
					<q-item
						clickable
						v-ripple
						dense
						@click="selectTemplate(item)"
						:active="isSelected(item)"
						class="item-web q-px-sm"
						active-class="border-color-yellow activeItem text-black"
						style="padding-top: 10px; padding-bottom: 10px"
					>
						<q-item-section side class="q-ml-xs q-pr-sm">
							<q-icon :name="showItemIcon(item.icon)" />
						</q-item-section>

						<q-item-section class="text-left text-body3 text-ink-1">{{
							item.toString()
						}}</q-item-section>
					</q-item>
				</div>
			</q-card-section>
		</div>

		<div class="s-dialog-plugin" v-else>
			<q-card-section class="q-pt-none q-px-none">
				<div class="text-left text-subtitle3 q-mb-sm text-ink-3">
					{{ t('vault_t.select_vault') }}
				</div>
				<div class="row align-center justify-center">
					<q-select
						class="select-vault"
						popup-content-class="options_selected_Account"
						standout
						dense
						emit-value
						map-options
						behavior="menu"
						v-model="model"
						:menu-offset="[0, 4]"
						dropdown-icon="sym_r_expand_more"
						:options="
							vaults.map((v) => ({
								value: v.id,
								label: `${v.org?.name || ''}${v.org?.name ? ' / ' : ''}${
									v.name
								}`
							}))
						"
						@update:model-value="changeModel"
						style="width: 100%"
					>
						<template
							v-slot:option="{ itemProps, opt, selected, toggleOption }"
						>
							<q-item v-bind="itemProps">
								<q-item-section>
									<q-item-label class="text-body3">{{
										opt.label
									}}</q-item-label>
								</q-item-section>
								<q-item-section side>
									<q-checkbox
										:model-value="selected"
										checked-icon="sym_r_check_circle"
										unchecked-icon=""
										indeterminate-icon="help"
										@update:model-value="toggleOption(opt.label)"
									/>
								</q-item-section>
							</q-item>
						</template>
					</q-select>
				</div>
			</q-card-section>

			<div
				class="text-subtitle3 text-ink-3 text-left q-px-none q-ml-xs q-mb-xs"
			>
				{{ t('vault_t.what_kind_of_item_you_would_like_to_add') }}
			</div>
			<q-card-section class="row q-gutter-md q-px-none" style="padding-top: 0">
				<div
					v-for="(item, index) in templates"
					:key="index"
					style="width: calc(50% - 12px)"
				>
					<q-item
						clickable
						v-ripple
						@click="selectTemplate(item)"
						:active="isSelected(item)"
						class="item-web q-px-xs q-py-sm"
						active-class="border-color-yellow activeItem text-black"
					>
						<q-item-section side class="q-ml-xs q-pr-sm">
							<q-icon :name="showItemIcon(item.icon)" />
						</q-item-section>

						<q-item-section class="text-left text-body3 text-ink-1">
							{{ item.toString() }}
						</q-item-section>
					</q-item>
				</div>
			</q-card-section>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import {
	FieldType,
	cloneItemTemplates,
	ItemTemplate,
	Vault
} from '@didvault/sdk/src/core';
import { getAppPlatform } from '../../../application/platform';
import { ExtensionPlatform } from '../../../platform/bex/front/platform';
import { app } from '../../../globals';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	option: Object
});

const vaults = ref([]);
const model = ref(props.option!.value);
const CustomRef = ref();

const { t } = useI18n();

const isWeb = ref(
	process.env.APPLICATION == 'VAULT' || process.env.PLATFORM == 'DESKTOP'
);

let templates: ItemTemplate[] = cloneItemTemplates();

templates = templates.filter(
	(item) => item.id !== 'crypto' && item.id !== 'exchange' && item.id != 'vc'
);

let selectedTemplate = ref(templates[0]);

async function selectTemplate(template: ItemTemplate) {
	// unlock
	selectedTemplate.value = template;
}

const isSelected = (template: ItemTemplate) => {
	return selectedTemplate.value.toString() == template.toString();
};

async function onOKClick() {
	try {
		if (process.env.PLATFORM == 'BEX') {
			if (selectedTemplate.value.icon == 'web') {
				const url = selectedTemplate.value.fields.find((it) => {
					return it.type == FieldType.Url;
				});
				const tab = await (
					getAppPlatform() as unknown as ExtensionPlatform
				).getCurrentTab();
				if (url && tab && tab.url) {
					try {
						const urlObj = new URL(tab.url);
						const baseUrl = urlObj.origin;
						url.value = baseUrl;
					} catch (error) {
						url.value = tab.url;
					}
				}
			}
		}

		const hasVault = vaults.value.find((c) => c.id === model.value);

		CustomRef.value.onDialogOK({
			selectedTemplate: selectedTemplate.value,
			vault: hasVault
		});
	} catch (e) {
		console.error(e);
	}
}

const changeModel = (value) => {
	model.value = value;
};

const showItemIcon = (name: string) => {
	switch (name) {
		case 'vault':
			return 'sym_r_language';
		case 'web':
			return 'sym_r_language';
		case 'computer':
			return 'sym_r_computer';
		case 'creditCard':
			return 'sym_r_credit_card';
		case 'bank':
			return 'sym_r_account_balance';
		case 'wifi':
			return 'sym_r_wifi_password';
		case 'passport':
			return 'sym_r_assignment_ind';
		case 'authenticator':
			return 'sym_r_password';
		case 'document':
			return 'sym_r_list_alt';
		case 'custom':
			return 'sym_r_chrome_reader_mode';

		default:
			break;
	}
};

const moveMineToFront = (array: Vault[], key: string, value: any) => {
	const index = array.findIndex((obj) => obj[key] === value);
	if (index > -1) {
		const obj = array.splice(index, 1)[0];
		array.unshift(obj);
	}
	return array;
};

onMounted(async () => {
	vaults.value = await moveMineToFront(app.vaults, 'id', props.option?.value);
});
</script>

<style lang="scss" scoped>
.select-vault {
	::v-deep(.q-field__control) {
		background: $background-1 !important;
		color: $ink-2;
	}

	::v-deep(.q-field__native) {
		color: $ink-2;
	}
}
.s-dialog-plugin-web {
	width: 100%;
	border-radius: 12px;

	.select-vault {
		border: 1px solid $input-stroke;
		border-radius: 8px;
		overflow: hidden;
	}

	.item-web {
		border-radius: 8px;
		border: 1px solid $btn-stroke;
	}

	.but-creat-web {
		border-radius: 8px;
		background: $yellow-default;
	}

	.but-cancel-web {
		border-radius: 8px;
		border: 1px solid $btn-stroke;
	}
}

.s-dialog-plugin {
	// width: 400px;
	border-radius: 12px;

	.select-vault {
		border: 1px solid $input-stroke;
		border-radius: 8px;
		overflow: hidden;
	}

	.item-web {
		border-radius: 8px;
		border: 1px solid $btn-stroke;
	}

	.title {
		text-align: center;
	}
}

.but-creat {
	border-radius: 8px;
	height: 48px;
}

.but-cancel2 {
	height: 48px;
	border-radius: 8px;
	border: 1px solid $separator;
}
</style>

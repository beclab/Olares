<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('vault_t.move_item_to')"
		:ok="t('vault_t.move_item')"
		:cancel="t('cancel')"
		:loading="moving"
		size="medium"
		@onSubmit="handleOk"
	>
		<q-card-section class="q-pt-md q-px-none">
			<q-select
				class="move-vault"
				popup-content-class="options_selected_Account"
				standout
				dense
				emit-value
				map-options
				popup-no-route-dismiss
				behavior="menu"
				v-model="model"
				:menu-offset="[0, 4]"
				dropdown-icon="sym_r_expand_more"
				:options="
					vaults.map((v) => ({
						value: v.id,
						label: v.name,
						disable: !app.isEditable(v)
					}))
				"
				@update:model-value="changeModel"
				style="width: 100%"
			>
				<template v-slot:option="{ itemProps, opt, selected, toggleOption }">
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
		</q-card-section>
	</bt-custom-dialog>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Vault } from '@didvault/sdk/src/core';
import { app } from '../../../globals';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	selected: Array,
	leftText: String,
	rightText: String
});

const model = ref();
const items = ref();
const vaults = ref<any[]>([]);
const vaultSelect = ref();
const moving = ref(false);

const { t } = useI18n();

const changeModel = (value) => {
	vaultSelect.value = vaults.value.find((item) => item.id === value);
};

const handleOk = async () => {
	moving.value = true;
	try {
		await app.moveItems(
			items.value.map((i: { item: any }) => i.item),
			vaultSelect.value!
		);
		moving.value = false;
		onDialogOK();
	} catch (e) {
		moving.value = false;
		console.error('catch', e);
	}
};

onMounted(() => {
	items.value = props.selected;
	const sourceVaults = items.value.reduce(
		(sv, i) => sv.add(i.vault),
		new Set<Vault>()
	);

	vaults.value =
		sourceVaults.size === 1
			? app.vaults.filter(
					(v) =>
						app.hasWritePermissions(v) &&
						v !== sourceVaults.values().next().value
			  )
			: app.vaults.filter((v) => app.hasWritePermissions(v));
});

const CustomRef = ref();

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style scoped lang="scss">
.move-vault {
	border: 1px solid $input-stroke;
	border-radius: 8px;
	overflow: hidden;

	::v-deep(.q-field__control) {
		background: $background-1 !important;
		color: $ink-2;
	}

	::v-deep(.q-field__native) {
		color: $ink-2;
	}
}
.d-creatVault {
	.q-dialog-plugin {
		width: 400px;
		border-radius: 12px;

		.confirm {
			padding: 8px 12px;
		}
		.reset {
			padding: 8px 12px;
			border: 1px solid $separator;
		}
	}
}
</style>

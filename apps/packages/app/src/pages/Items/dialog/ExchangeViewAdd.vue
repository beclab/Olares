<template>
	<!-- notice dialogRef here -->
	<q-dialog
		class="d-creatVault text-center"
		ref="dialogRef"
		@hide="onDialogHide"
	>
		<q-card class="q-dialog-plugin">
			<q-card-section class="q-py-xs">
				<div class="text-h5 q-my-lg">{{ t('new_exchange') }}</div>
			</q-card-section>

			<q-form @submit="onOKClick" class="q-pa-md form">
				<q-select
					class="q-mb-md exchange-style"
					v-model="exchange"
					borderless
					:options="exchanges"
					autofocus
					popup-content-style="display: block"
					label="Exchange"
				>
					<template v-slot:selected>
						<div class="row items-center full-width color-black">
							<div>{{ exchange.name }}</div>
						</div>
					</template>
					<template v-slot:option="scope">
						<q-item
							v-bind="scope.itemProps"
							dense
							class="dialog-option row items-center"
						>
							<q-icon v-if="scope.opt.name === exchange.name" name="check" />
							<span>{{ scope.opt.name }}</span>
						</q-item>
					</template>
				</q-select>

				<div
					class="q-mb-md"
					v-for="(field, index) in exchange.fields"
					:key="'fa' + index"
				>
					<q-input
						borderless
						v-model="field.value"
						:label="field.name"
						v-if="index > 0"
						class="exchange-input-style"
					/>
				</div>

				<div class="row q-mb-lg">
					<q-checkbox
						v-model="show_name_in_sign_sever"
						:label="t('vault_t.show_in_sign_server')"
					/>
					<q-checkbox
						v-model="sign_in_server"
						:label="t('vault_t.can_signing')"
						:disable="!show_name_in_sign_sever"
					/>
				</div>

				<div class="row justify-between q-mb-md">
					<q-btn
						clickable
						v-ripple
						class="bg-blue text-body1 but-creat q-mx-xs q-px-md items-center justify-center"
						type="submit"
					>
						{{ t('create') }}
					</q-btn>

					<q-item
						clickable
						v-ripple
						class="bg-grey-11 text-ink-2 text-body1 but-cancel q-mx-xs q-px-md items-center justify-center"
						@click="onCancelClick"
					>
						{{ t('vault_t.discard') }}
					</q-item>
				</div>
			</q-form>
		</q-card>
	</q-dialog>
</template>

<script lang="ts">
import { ref, watch } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import {
	FIELD_DEFS,
	ExchangeTemplate,
	EXCHANGE_TEMPLATES
} from '@didvault/sdk/src/core';

import {
	SecertType,
	HostType,
	Secert,
	SecertData
} from '@didvault/sdk/src/core';
import { app } from '../../../globals';
import { AddSecertParams } from '@didvault/sdk/src/core/api';

import { useI18n } from 'vue-i18n';

export default {
	name: 'ExchangeViewAdd',

	emits: [...useDialogPluginComponent.emits],

	setup() {
		let exchanges: ExchangeTemplate[] = [];
		for (let ex of EXCHANGE_TEMPLATES) {
			exchanges.push({
				name: ex.name,
				icon: ex.icon,
				fields: ex.fields.map((f: any) => ({
					type: f.type,
					name: f.name
				}))
			});
		}

		let exchange = ref(exchanges[0]);

		const { dialogRef, onDialogHide, onDialogOK, onDialogCancel } =
			useDialogPluginComponent();

		let show_name_in_sign_sever = ref(true);
		let sign_in_server = ref(true);

		const { t } = useI18n();

		watch(
			() => show_name_in_sign_sever.value,
			(newVal) => {
				if (!newVal) {
					sign_in_server.value = false;
				}
			}
		);

		async function onOKClick() {
			try {
				let result = exchange.value;
				result.fields[0].value = exchange.value.name;
				result.fields = result.fields.filter(function (e) {
					if (e.name == t('subaccount')) {
						if (e.value) {
							return true;
						}
						return false;
					}
					return true;
				});

				const nameFailed = result.fields.find((e) => e.name == t('account'));
				const subNameFailed = result.fields.find(
					(e) => e.name == t('subaccount')
				);

				try {
					let d = {};

					const secert: Secert = new Secert({
						secert_type: SecertType.Exchange,
						host_type: HostType.None,
						name: subNameFailed
							? subNameFailed.value
							: nameFailed
							? nameFailed.value
							: ''
					});

					secert.host_type = HostType.None;
					if (show_name_in_sign_sever.value) {
						secert.host_type = HostType.ONLY_SAVE_NAME;
					}

					if (sign_in_server.value) {
						secert.host_type = HostType.SAVE_SECERT;
					}

					for (const l of result.fields) {
						if (secert.host_type != HostType.SAVE_SECERT) {
							if (
								l.name.toLocaleLowerCase() == 'key' ||
								l.name.toLocaleLowerCase() == 'secret' ||
								l.name.toLocaleLowerCase() == 'passphrase'
							) {
								continue;
							}
						}
						d[l.name.toLocaleLowerCase()] = l.value;
					}
					secert.data = new SecertData(d);

					if (
						secert.host_type == HostType.ONLY_SAVE_NAME ||
						secert.host_type == HostType.SAVE_SECERT
					) {
						await app.addSecertToSignServer(
							new AddSecertParams({
								secert
							})
						);
					}
				} catch (err) {
					console.error(err);
				}

				onDialogOK(exchange.value);
			} catch (e) {
				console.error(e);
			}
		}

		return {
			show_name_in_sign_sever,
			sign_in_server,
			dialogRef,
			FIELD_DEFS,
			onDialogHide,
			onOKClick,
			onCancelClick: onDialogCancel,
			exchanges,
			exchange,
			t
		};
	}
};
</script>

<style lang="scss" scoped>
.d-creatVault {
	.q-dialog-plugin {
		width: 500px;
		border-radius: 20px;
	}

	.form {
		padding-left: 20px;
		padding-right: 20px;
	}

	.but-creat {
		width: 46%;
		border-radius: 10px;

		text-transform: capitalize;
	}

	.but-cancel {
		width: 46%;
		border-radius: 10px;

		text-transform: capitalize;
	}

	.exchange-style {
		border-radius: 10px;
		border: 1px solid $separator;
		padding: 0 20px;
	}

	.exchange-input-style {
		border-radius: 10px;
		border: 1px solid $separator;
		border-radius: 10px;
		padding-left: 20px;
	}
}
</style>

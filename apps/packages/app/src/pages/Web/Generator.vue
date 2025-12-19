<template>
	<q-page class="generrator">
		<div
			class="text-subtitle2 text-ink-1 row items-center justify-start q-pl-sm title"
		>
			<q-icon
				v-if="platform === 'MOBILE'"
				name="sym_r_chevron_left"
				size="24px"
				@click="goBack"
			/>
			<q-icon name="sym_r_casino" size="24px" />

			<span class="q-ml-sm">{{ t('password_generators') }}</span>
		</div>

		<div class="content">
			<div class="box">
				<div class="box-toggle">
					<q-btn
						class="passphrase text-ink-2 btn"
						:class="models === 'passphrase' ? 'text-ink-1 active' : ''"
						:label="t('passphrase')"
						@click="toggle('passphrase')"
					/>
					<q-btn
						class="random text-ink-2 btn"
						:class="models === 'random' ? 'text-ink-1 active' : ''"
						:label="t('random_string')"
						@click="toggle('random')"
					/>
				</div>

				<div class="separator" v-if="models === 'passphrase'">
					<q-select
						class="separator-select"
						borderless
						dense
						emit-value
						map-options
						behavior="menu"
						v-model="separatorModel"
						:options="separatorOptions"
						:label="t('word_separator')"
						dropdown-icon="sym_r_expand_more"
						@update:model-value="generate"
					/>
				</div>

				<div class="separator" v-if="models === 'passphrase'">
					<q-select
						class="separator-select"
						borderless
						dense
						emit-value
						map-options
						behavior="menu"
						v-model="languageModel"
						:options="AVAILABLE_LANGUAGES"
						:label="t('language')"
						dropdown-icon="sym_r_expand_more"
						@update:model-value="generate"
					/>
				</div>

				<div
					class="words row items-center justify-between"
					v-if="models === 'passphrase'"
				>
					<q-slider
						v-model="wordsModel"
						:min="3"
						:max="12"
						:step="1"
						snap
						color="yellow"
						style="width: 80%"
						@change="changeWords"
					/>
					<span> {{ wordsModel }} {{ t('words') }} </span>
				</div>

				<div class="wordsPanel bg-background-3" v-if="models === 'passphrase'">
					{{ wordPanel }}
				</div>

				<div class="lowercase" v-if="models === 'random'">
					<q-item tag="label" dense v-ripple>
						<q-item-section>
							<q-item-label>a-z</q-item-label>
						</q-item-section>
						<q-item-section avatar>
							<bt-switch
								size="sm"
								truthy-track-color="yellow-default"
								v-model="lowercaseModel"
								@update:model-value="generate"
							/>
						</q-item-section>
					</q-item>
				</div>

				<div class="capital" v-if="models === 'random'">
					<q-item tag="label" dense v-ripple>
						<q-item-section>
							<q-item-label>A-Z</q-item-label>
						</q-item-section>
						<q-item-section avatar>
							<bt-switch
								size="sm"
								truthy-track-color="yellow-default"
								v-model="capitalModel"
								@update:model-value="generate"
							/>
						</q-item-section>
					</q-item>
				</div>

				<div class="capital" v-if="models === 'random'">
					<q-item tag="label" dense v-ripple>
						<q-item-section>
							<q-item-label>0-9</q-item-label>
						</q-item-section>
						<q-item-section avatar>
							<bt-switch
								size="sm"
								truthy-track-color="yellow-default"
								v-model="numberModel"
								@update:model-value="generate"
							/>
						</q-item-section>
					</q-item>
				</div>

				<div class="capital" v-if="models === 'random'">
					<q-item tag="label" dense v-ripple>
						<q-item-section>
							<q-item-label>?()/%...</q-item-label>
						</q-item-section>
						<q-item-section avatar>
							<bt-switch
								size="sm"
								truthy-track-color="yellow-default"
								v-model="otherModel"
								@update:model-value="generate"
							/>
						</q-item-section>
					</q-item>
				</div>

				<div
					class="math row items-center justify-between"
					v-if="models === 'random'"
				>
					<span>length</span>
					<q-slider
						v-model="mathModel"
						:min="5"
						:max="150"
						:step="1"
						snap
						color="yellow"
						style="width: 75%"
					/>
					<span>{{ mathModel }}</span>
				</div>

				<div class="mathPanel bg-background-3" v-if="models === 'random'">
					{{ chartPanel }}
				</div>

				<div class="operation row items-center justify-center">
					<q-btn
						borderless
						class="regenerate text-ink-1 text-caption"
						:label="t('regenerate')"
						icon="autorenew"
						@click="generate"
					/>
					<q-btn
						borderless
						class="text-ink-1 text-caption copy"
						:label="t('copy')"
						icon="content_copy"
						@click="copy"
					/>
				</div>
			</div>
		</div>
	</q-page>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { app } from '../../globals';
import { useI18n } from 'vue-i18n';

import { generatePassphrase } from '../../passphrase/diceware';

import { AVAILABLE_LANGUAGES } from '../../passphrase/wordlists';

import { randomString, chars, getPlatform } from '@didvault/sdk/src/core';
import { notifyFailed, notifySuccess } from '../../utils/notifyRedefinedUtil';

export default defineComponent({
	name: 'passwordGenerator',
	components: {},
	setup() {
		const models = ref('passphrase');
		const separatorModel = ref(' ');
		const separatorOptions = ref([
			{
				value: '-',
				label: 'Dash ( - )'
			},
			{
				value: '_',
				label: 'Underscore ( _ )'
			},
			{
				value: '/',
				label: 'Slash ( / )'
			},
			{
				value: ' ',
				label: 'Space (   )'
			}
		]);
		const Router = useRouter();
		const languageModel = ref('en');
		const wordsModel = ref(4);
		const lowercaseModel = ref(true);
		const capitalModel = ref(true);
		const numberModel = ref(true);
		const otherModel = ref(false);
		const mathModel = ref(20);
		const wordPanel = ref();
		const chartPanel = ref();
		const platform = ref(process.env.PLATFORM);

		const toggle = (btn: string) => {
			models.value = btn;
			generate();
		};

		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		const changeWords = (_value: string) => {
			generate();
		};

		const generate = async () => {
			const separator = separatorModel.value || '-';
			const language = languageModel.value || app.state.device.locale;
			let charSet = '';

			switch (models.value) {
				case 'passphrase':
					wordPanel.value = await generatePassphrase(
						wordsModel.value,
						separator,
						[language]
					);
					break;
				case 'random':
					lowercaseModel.value && (charSet += chars.lower);
					capitalModel.value && (charSet += chars.upper);
					numberModel.value && (charSet += chars.numbers);
					otherModel.value && (charSet += chars.other);
					chartPanel.value = charSet
						? await randomString(mathModel.value, charSet)
						: '';
					break;
			}
		};

		const copy = () => {
			let copyText = '';
			switch (models.value) {
				case 'passphrase':
					copyText = wordPanel.value;
					break;
				case 'random':
					copyText = chartPanel.value;
					break;
			}
			const platform = getPlatform();
			platform
				.setClipboard(copyText)
				.then(() => {
					// success!
					notifySuccess(t('copy_success'));
				})
				.catch(() => {
					notifyFailed(t('copy_fail'));
				});
		};

		const goBack = () => {
			Router.go(-1);
		};

		onMounted(() => {
			generate();
		});

		const { t } = useI18n();

		return {
			models,
			separatorModel,
			separatorOptions,
			languageModel,
			wordsModel,
			lowercaseModel,
			capitalModel,
			numberModel,
			otherModel,
			mathModel,
			AVAILABLE_LANGUAGES,
			wordPanel,
			chartPanel,

			toggle,
			generate,
			copy,
			goBack,
			platform,
			changeWords,
			t
		};
	}
});
</script>

<style lang="scss" scoped>
.generrator {
	height: 100%;
	display: flex;
	flex-direction: column;

	.title {
		padding-top: 20px;
	}

	.content {
		height: 90%;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		margin: 0 16px;

		.box {
			width: 100%;
			max-width: 30em;
			border: 1px solid $separator;
			border-radius: 20px;
			padding-bottom: 20px;

			.q-btn:before {
				box-shadow: none;
			}

			.box-toggle {
				width: 90%;
				height: 56px;
				display: flex;
				align-items: center;
				justify-content: space-between;
				margin: 20px auto 0;

				.btn {
					width: 45%;
					border: 1px solid $separator;
					border-radius: 10px;
					padding-left: 0;
					padding-right: 0;

					&.active {
						border: 1px solid $yellow-default;
					}
				}
			}

			.separator {
				width: 90%;
				margin: 20px auto 0;
				.separator-select {
					border: 1px solid $separator;
					padding: 0 8px;
					border-radius: 8px;
					overflow: hidden;
				}
			}

			.words {
				width: 86%;
				margin: 20px auto 0;
			}

			.lowercase {
				width: 90%;
				margin: 12px auto 0;
				border-radius: 10px;
				overflow: hidden;
				border: 1px solid $separator;
			}

			.capital {
				width: 90%;
				margin: 12px auto 0;
				border-radius: 10px;
				overflow: hidden;
				border: 1px solid $separator;
			}

			.math {
				width: 86%;
				margin: 12px auto 0;
			}

			.mathPanel {
				width: 90%;
				margin: 12px auto 0;
				border-radius: 10px;
				letter-spacing: inherit;
				vertical-align: baseline;
				text-align: center;
				overflow-wrap: break-word;
				text-align: center;
				font-size: 120%;
				padding: 1.2em;
			}

			.wordsPanel {
				width: 90%;
				margin: 12px auto 0;
				border-radius: 10px;
				letter-spacing: inherit;
				vertical-align: baseline;
				text-align: center;
				overflow-wrap: break-word;
				text-align: center;
				font-size: 120%;
				padding: 1.2em;
			}

			.operation {
				width: 90%;
				margin: 10px auto 0;

				.regenerate {
					margin-right: 20px;
				}

				.regenerate,
				.copy {
					border: 1px solid $separator;
					border-radius: 10px;
				}
			}
		}
	}
}
</style>

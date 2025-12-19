<template>
	<div class="upgrade">
		<div class="logo-wrap">
			<img
				src="../../../assets/desktop/upgrade-success.svg"
				v-if="['completed', 'not_running'].includes(upgradeStore.state)"
				alt="success"
			/>
			<img
				src="../../../assets/desktop/upgrade-fail.svg"
				v-else-if="upgradeStore.state === 'failed'"
				alt="fail"
			/>
			<div class="logo" v-else>
				<div class="shadow"></div>
				<img src="../../../assets/desktop/upgrading.svg" alt="upgrading" />
			</div>
		</div>

		<div
			class="content"
			v-if="['completed', 'not_running'].includes(upgradeStore.state)"
		>
			<div>{{ t('upgrade.success_title') }}</div>
			<div>{{ t('upgrade.success_message') }}</div>
		</div>

		<div class="content" v-else-if="upgradeStore.state === 'failed'">
			<div>{{ t('upgrade.fail_title') }}</div>
			<div>{{ t('upgrade.fail_message') }}</div>
		</div>

		<div class="content" v-else>
			<div>{{ t('upgrade.running_title') }}</div>
			<div>{{ t('upgrade.running_message') }}</div>
		</div>

		<div
			class="confirm"
			v-if="!['started', 'running'].includes(upgradeStore.state)"
			@click="confirm"
		>
			{{ t('buttons.confirm') }}
		</div>
		<div class="confirm" v-else style="opacity: 0"></div>
	</div>
</template>
<script lang="ts" setup>
import { useUpgradeStore } from '../../../stores/desktop/upgrade';
import { useI18n } from 'vue-i18n';

const emits = defineEmits(['closeUpgrade']);

const upgradeStore = useUpgradeStore();
const { t } = useI18n();

const confirm = () => {
	emits('closeUpgrade');
};
</script>

<style lang="scss" scoped>
.upgrade {
	width: 100vw;
	height: 100vh;
	z-index: 100000;
	background: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(15px);
	position: fixed;
	top: 0;
	left: 0;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	.logo-wrap {
		width: 80px;
		height: 80px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.logo {
		width: 80px;
		height: 80px;
		position: relative;
		.shadow {
			width: 100%;
			height: 100%;
			// border: 4px solid transparent;
			// border-top-color: #ffffff;
			box-sizing: border-box;
			border-radius: 50%;
			animation: 2.5s loaderAni linear infinite;
			position: absolute;
			left: 0;
			right: 0;
			top: 0;
			bottom: 0;
			margin: auto;

			background-image: url('../../../assets/desktop/upgrading.png');
			background-repeat: no-repeat;
			background-size: 100%;
			background-position: center center;

			// &:before {
			// 	content: '';
			// 	display: block;
			// 	width: calc(100% + 0px);
			// 	height: calc(100% + 0px);
			// 	position: absolute;
			// 	left: -1px;
			// 	right: 0;
			// 	top: 0;
			// 	bottom: 0;
			// 	margin: auto;
			// 	border: 1px solid #ffffff;
			// 	box-sizing: content-box;
			// 	border-radius: 50%;
			// 	opacity: 0.5;
			// }

			@keyframes loaderAni {
				0% {
					transform: rotate(0deg);
				}
				100% {
					transform: rotate(360deg);
				}
			}
		}

		img {
			width: 58px;
			height: 58px;
			position: absolute;
			left: 0;
			right: 0;
			top: 0;
			bottom: 0;
			margin: auto;
		}
	}

	.content {
		margin-top: 16px;
		> :first-child {
			line-height: 32px;
			font-weight: 800;
			font-size: 20px;
			line-height: 24px;
			text-align: center;
			letter-spacing: 0.02em;
			color: #ffffff;
		}
		> :last-child {
			line-height: 24px;
			font-weight: 400;
			font-size: 12px;
			line-height: 16px;
			text-align: center;
			color: #ffffff;
			margin-top: 8px;
		}
	}
	.confirm {
		font-weight: 600;
		font-size: 12px;
		line-height: 16px;
		color: #ffffff;
		padding: 8px 32px;
		background: linear-gradient(90deg, #ffa337 0%, #ef438b 100%);
		border-radius: 16px;
		margin-top: 24px;
		cursor: pointer;
	}
}
</style>

<template>
	<div class="wrap">
		<div class="container">
			<div class="header">
				<div class="text-h3 text-ink-1">{{ t('home_welcome') }}</div>
				<div class="text-subtitle1 text-ink-3 q-mt-sm subtitle">
					{{ t('home_desc') }}
				</div>
			</div>
			<div class="row">
				<div class="col-6 q-pr-lg" style="padding-right: 40px">
					<div class="text-h5 text-ink-1">{{ t('home_start') }}</div>

					<div
						class="text-subtitle1 text-white create-btn create-application-btn row items-center justify-start"
						@click="onCreate"
					>
						<span class="add_wrap">
							<q-icon name="sym_r_add" size="20px" color="green" />
						</span>
						<span> {{ t('home_create') }} </span>
					</div>
					<input
						ref="uploadInput"
						type="file"
						style="display: none"
						accept=".zip,.rar,.7z,.tar,.tgz"
						@change="uploadFile"
					/>
					<div
						class="text-subtitle1 text-ink-1 create-btn create-con-btn row items-center justify-start"
						@click="uploadChat"
					>
						<span class="add_wrap bg-background-3">
							<q-icon name="sym_r_upgrade" size="20px" color="grey-5" />
						</span>
						<span> {{ t('home_update') }}</span>
					</div>
				</div>

				<div class="col-6" style="padding-left: 40px">
					<div class="text-h5 text-ink-1">{{ t('home_documents') }}</div>
					<document-link
						v-for="(item, index) in menuStore.documentList"
						:key="index"
						:data="item"
					></document-link>
					<div class="link-more text-body1 text-ink-2">
						{{ t('home_visit_1') }}
						<a
							href="https://docs.olares.com/zh/manual/overview.html"
							target="_blank"
							>docs.olares.com</a
						>
						{{ t('home_visit_2') }}
					</div>
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import axios from 'axios';
import { useRoute, useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useDevelopingApps } from '../stores/app';
import { useMenuStore } from '../stores/menu';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { pushToSystem } from './../utils/utils';

import DocumentLink from '../components/common/DocumentLink.vue';
import CreateAppName from '../components/dialog/CreateAppName.vue';
import { useDockerStore } from '@apps/studio/src/stores/docker';

const { t } = useI18n();
const $q = useQuasar();
const router = useRouter();
const route = useRoute();
const store = useDevelopingApps();
const menuStore = useMenuStore();
const dockerStore = useDockerStore();

const uploadInput = ref();

onMounted(async () => {
	//
});

function onCreate() {
	$q.dialog({
		component: CreateAppName
	});
}

const uploadChat = () => {
	uploadInput.value.click();
};

async function uploadFile(event: any) {
	console.log(event);
	const file = event.target.files[0];
	if (file) {
		const { status, message, appId } = await upload_dev_file(file);
		if (status) {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: message,
				notify_id: appId
			});
			if (appId) {
				await store.getApps();
				await menuStore.updateApplications();

				const current_app_name = store.apps.find(
					(app) => app.id === appId
				).appName;
				$q.loading.show({
					message: t('installing')
				});

				try {
					const res = await dockerStore.install_app(current_app_name);
					const namespace = res.namespace;
					$q.loading.hide();

					menuStore.currentItem = '/app/' + current_app_name;
					updateStatus(current_app_name, namespace);
				} catch (error) {
					$q.loading.hide();
				}
			}
		}
	} else {
		console.log('file selected failure');
	}
}

const updateStatus = async (app_name: string, namespace: string) => {
	await pushToSystem(app_name, router, namespace);
	await dockerStore.get_app_status(app_name);
	await dockerStore.get_app_install_state(app_name);
};

async function upload_dev_file(
	file: any
): Promise<{ status: boolean; message: string; appId?: number }> {
	try {
		const formData = new FormData();
		formData.append('chart', file);
		const res: any = await axios.post(
			store.url + '/api/command/upload-app-archive',
			formData,
			{
				headers: { 'Content-Type': 'multipart/form-data' }
			}
		);

		return {
			status: res.appId ? true : false,
			message: res.appId ? 'upload chart success' : res.data?.message,
			appId: res.appId
		};
	} catch (e: any) {
		return { status: false, message: e.message };
	}
}
</script>

<style lang="scss" scoped>
.wrap {
	width: 100%;
	height: 100vh;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0 44px;
	.container {
		width: 100%;
		max-width: 1280px;
		height: 100vh;
		.header {
			margin-top: 56px;
			margin-bottom: 20px;
			padding-bottom: 32px;
			border-bottom: 1px solid $separator;
		}
		.create-btn {
			width: 100%;
			height: 72px;
			border-radius: 12px;
			padding: 0 20px;
			box-sizing: border-box;
			cursor: pointer;
			.add_wrap {
				width: 32px;
				height: 32px;
				line-height: 32px;
				text-align: center;
				border-radius: 16px;
				margin-right: 12px;
				display: flex;
				align-items: center;
				justify-content: center;
			}

			.app-content {
				img {
					width: 32px;
					height: 32px;
					border-radius: 8px;
					margin-right: 12px;
				}
			}
			.app-status {
				i {
					width: 12px;
					height: 12px;
					display: inline-block;
					border-radius: 6px;
					background-color: green;
					margin-right: 8px;
				}
			}
		}
		.create-application-btn {
			margin-top: 12px;
			background-color: rgba(0, 190, 158, 1);
			.add_wrap {
				background-color: white;
			}
		}
		.create-con-btn {
			margin-top: 20px;
			border: 1px solid $separator;
			margin-bottom: 32px;
			.add_wrap {
				background-color: rgba(246, 246, 246, 1);
			}
		}
		.create-app-btn {
			margin-top: 20px;
			background-color: white;
			border: 1px solid rgba(235, 235, 235, 1);
			.add_wrap {
				background-color: rgba(246, 246, 246, 1);
			}
		}

		.link-more {
			margin-top: 32px;
			a {
				color: #00967d;
				text-decoration: none;
			}
		}
	}
}
</style>

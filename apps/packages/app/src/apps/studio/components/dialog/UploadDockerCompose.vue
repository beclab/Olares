<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('home_upload_compose')"
		:cancel="t('btn_cancel')"
		:ok="t('btn_confirm')"
		:okLoading="loading ? $t('loading') : false"
		size="large"
		@onSubmit="submit"
	>
		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('home_appname') }} <span class="text-red-default">*</span>
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					ref="appNameRef"
					dense
					borderless
					no-error-icon
					v-model="appTitle"
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.appNameRules.placeholder"
					:rules="ruleConfig.appNameRules.rules"
				>
				</q-input>
			</div>
		</div>

		<div class="form-item row q-mt-md">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('upload') }} <span class="text-red-default">*</span>
			</div>
			<div class="form-item-value q-mb-lg">
				<input
					ref="fileInput"
					type="file"
					style="display: none"
					accept=".yml"
					@change="handleFileChange"
				/>
				<div
					class="file-upload-area"
					:class="{ 'has-file': selectedFile, 'has-error': fileError }"
					@click="fileInput.click()"
				>
					<div v-if="!selectedFile" class="text-ink-3 upload-prompt">
						<q-icon name="sym_r_upload_file" size="20px" class="q-mr-sm" />
						{{ t('home_compose_file_hint') }}
					</div>
					<div v-else class="file-info row items-center justify-between">
						<div class="row items-center text-ink-1">
							<q-icon name="sym_r_description" size="20px" class="q-mr-sm" />
							<span>{{ selectedFile.name }}</span>
						</div>
						<q-btn
							flat
							dense
							round
							size="sm"
							icon="sym_r_close"
							color="ink-2"
							@click.stop="removeFile"
							class="delete-btn"
						>
							<q-tooltip>{{ t('btn_delete') }}</q-tooltip>
						</q-btn>
					</div>
				</div>
				<div
					v-if="fileError"
					class="text-caption text-negative q-mt-sm"
					style="font-size: 11px"
				>
					{{ t('home_compose_file_required') }}
				</div>

				<!-- 帮助说明区域 -->
				<div class="help-section q-mt-md">
					<div class="info-banner">
						<q-icon
							name="sym_r_info"
							size="18px"
							class="q-mr-sm"
							color="primary"
						/>
						<div class="info-content">
							<div class="text-body2 text-ink-1 q-mb-xs">
								{{ t('home_compose_entrance_tip') }}
								<q-tooltip anchor="top middle" self="bottom middle">
									{{ t('home_compose_entrance_why') }}
								</q-tooltip>
							</div>

							<div class="label-example q-mb-xs">
								<code class="label-code">
									labels:<br />
									&nbsp;&nbsp;olares.service.type: Entrance
								</code>
							</div>

							<!-- 可折叠的示例 -->
							<q-expansion-item
								v-model="showExample"
								:label="t('home_compose_entrance_example_title')"
								class="example-expansion"
								dense
								header-class="text-caption text-primary"
							>
								<div class="example-code q-mt-sm">
									<div class="code-container">
										<div class="code-header">
											<span class="text-caption text-ink-2"
												>docker-compose.yml</span
											>
											<div class="action-buttons">
												<q-btn
													flat
													dense
													size="sm"
													icon="sym_r_download"
													color="primary"
													class="action-btn"
													@click="downloadExampleCode"
												>
													<q-tooltip>{{ t('btn_download') }}</q-tooltip>
												</q-btn>
												<q-btn
													flat
													dense
													size="sm"
													icon="sym_r_content_copy"
													color="primary"
													class="action-btn"
													@click="copyExampleCode"
												>
													<q-tooltip>{{
														copied ? t('btn_copied') : t('btn_copy')
													}}</q-tooltip>
												</q-btn>
											</div>
										</div>
										<pre class="code-block"><code>services:
  web:
    container_name: web
    image: quay.io/kompose/web
    ports:
      - "8080:8080"
    environment:
      - GET_HOSTS_FROM=dns
    labels:
      olares.service.type: Entrance  <span class="code-comment-highlight"># {{ t('home_compose_entrance_label') }}</span>

  redis-leader:
    container_name: redis-leader
    image: redis
    ports:
      - "6379"

  redis-replica:
    container_name: redis-replica
    image: redis
    ports:
      - "6379"
    command: redis-server --replicaof redis-leader 6379 --dir /tmp</code></pre>
									</div>
								</div>
							</q-expansion-item>

							<div class="text-caption text-ink-3 q-mt-sm">
								<q-icon name="sym_r_help" size="14px" class="q-mr-xs" />
								{{ t('home_compose_entrance_note') }}
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import axios from 'axios';
import { ruleConfig } from './../../types/config';
import { useDockerStore } from './../../stores/docker';
import { useDevelopingApps } from '../../stores/app';
import { useMenuStore } from '../../stores/menu';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

const { t } = useI18n();
const router = useRouter();

const dockerStore = useDockerStore();
const appStores = useDevelopingApps();
const menuStore = useMenuStore();

const CustomRef = ref();
const appNameRef = ref();
const fileInput = ref();
const loading = ref(false);
const appTitle = ref();
const selectedFile = ref<File | null>(null);
const fileError = ref(false);
const showExample = ref(false);
const copied = ref(false);

const exampleCode = `services:
  web:
    container_name: web
    image: quay.io/kompose/web
    ports:
      - "8080:8080"
    environment:
      - GET_HOSTS_FROM=dns
    labels:
      olares.service.type: Entrance  # ${t('home_compose_entrance_label')}

  redis-leader:
    container_name: redis-leader
    image: redis
    ports:
      - "6379"

  redis-replica:
    container_name: redis-replica
    image: redis
    ports:
      - "6379"
    command: redis-server --replicaof redis-leader 6379 --dir /tmp`;

const handleFileChange = (event: any) => {
	const file = event.target.files[0];
	if (file) {
		selectedFile.value = file;
		fileError.value = false;
	}
};

const removeFile = () => {
	selectedFile.value = null;
	fileError.value = false;
	if (fileInput.value) {
		fileInput.value.value = '';
	}
};

const copyExampleCode = async () => {
	try {
		await navigator.clipboard.writeText(exampleCode);
		copied.value = true;
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.copy_success')
		});
		setTimeout(() => {
			copied.value = false;
		}, 2000);
	} catch (err) {
		console.error('Failed to copy:', err);
	}
};

const downloadExampleCode = () => {
	const blob = new Blob([exampleCode], { type: 'text/yaml;charset=utf-8' });
	const url = URL.createObjectURL(blob);
	const link = document.createElement('a');
	link.href = url;
	link.download = 'docker-compose-demo.yml';
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);

	URL.revokeObjectURL(url);

	BtNotify.show({
		type: NotifyDefinedType.SUCCESS,
		message: t('message.download_success')
	});
};

const submit = async () => {
	console.log('click submit');
	appNameRef.value.validate();

	if (appNameRef.value.hasError) return;

	if (!selectedFile.value) {
		fileError.value = true;
		return;
	}

	try {
		loading.value = true;

		const appName = appTitle.value;

		const formData = new FormData();
		formData.append('file', selectedFile.value);
		formData.append('title', appTitle.value);

		await axios.post(appStores.url + `/api/command/apps/kompose`, formData, {
			headers: { 'Content-Type': 'multipart/form-data' }
		});
		await appStores.getApps();

		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('message.upload_compose_success')
		});

		dockerStore.appStatus = undefined;
		router.push({ path: '/app/' + appName });
		menuStore.currentItem = '/app/' + appName;

		loading.value = false;
		CustomRef.value.onDialogOK();
	} catch (error: any) {
		loading.value = false;
	}
};

async function getAppName(title: string) {
	const current_app = appStores.apps.find((item) => item.title == title);
	if (!current_app) {
		throw new Error('Application not found');
	}
	return current_app.appName;
}
</script>

<style lang="scss" scoped>
.form-item {
	.form-item-key {
		width: 120px;
		height: 40px;
		line-height: 40px;
		text-align: center;
	}
	.form-item-value {
		flex: 1;
	}
}

.file-upload-area {
	border: 1px dashed $separator;
	border-radius: 8px;
	padding: 12px 16px;
	cursor: pointer;
	transition: all 0.3s;
	background-color: $background-2;

	&:hover {
		border-color: $primary;
		background-color: rgba(0, 190, 158, 0.05);
	}

	&.has-file {
		border-style: solid;
	}

	&.has-error {
		border-color: $red-default;
	}

	.upload-prompt {
		display: flex;
		align-items: center;
	}

	.file-info {
		width: 100%;

		.delete-btn {
			opacity: 0.6;
			transition: opacity 0.2s;

			&:hover {
				opacity: 1;
				background-color: rgba(0, 0, 0, 0.05);
			}
		}
	}
}

.help-section {
	.info-banner {
		display: flex;
		background-color: rgba(0, 190, 158, 0.08);
		border: 1px solid rgba(0, 190, 158, 0.2);
		border-radius: 8px;
		padding: 16px;

		.info-content {
			flex: 1;
		}

		.label-example {
			display: flex;
			align-items: flex-start;
			margin: 8px 0;

			.label-code {
				background-color: rgba(0, 0, 0, 0.05);
				padding: 8px 12px;
				border-radius: 4px;
				font-family: 'Courier New', monospace;
				font-size: 12px;
				line-height: 1.5;
				color: $ink-1;
				white-space: pre;
			}
		}

		.example-expansion {
			background-color: transparent;
			border: none;

			:deep(.q-expansion-item__container) {
				border: none;
			}
		}

		.example-code {
			.code-container {
				position: relative;
				background-color: #1e1e1e;
				border-radius: 8px;
				overflow: hidden;

				.code-header {
					display: flex;
					justify-content: space-between;
					align-items: center;
					padding: 8px 12px;
					background-color: rgba(0, 0, 0, 0.2);
					border-bottom: 1px solid rgba(255, 255, 255, 0.1);

					.action-buttons {
						display: flex;
						gap: 4px;
					}

					.action-btn {
						min-height: 24px;
						padding: 4px 8px;
						transition: all 0.2s;

						&:hover {
							background-color: rgba(255, 255, 255, 0.1);
						}
					}
				}

				.code-block {
					background-color: #1e1e1e;
					border: none;
					border-radius: 0;
					padding: 16px;
					margin: 0;
					overflow-x: auto;
					max-height: 400px;

					&::-webkit-scrollbar {
						width: 8px;
						height: 8px;
					}

					&::-webkit-scrollbar-track {
						background: rgba(255, 255, 255, 0.05);
					}

					&::-webkit-scrollbar-thumb {
						background: rgba(255, 255, 255, 0.2);
						border-radius: 4px;

						&:hover {
							background: rgba(255, 255, 255, 0.3);
						}
					}

					code {
						font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
						font-size: 13px;
						line-height: 1.6;
						color: #d4d4d4;
						white-space: pre;
						display: block;

						.code-comment-highlight {
							color: #6a9955;
							background-color: rgba(106, 153, 85, 0.15);
							padding: 2px 6px;
							border-radius: 3px;
							font-weight: 500;
							box-shadow: 0 0 8px rgba(106, 153, 85, 0.3);
							animation: highlight-pulse 2s ease-in-out infinite;
						}

						@keyframes highlight-pulse {
							0%,
							100% {
								box-shadow: 0 0 8px rgba(106, 153, 85, 0.3);
							}
							50% {
								box-shadow: 0 0 12px rgba(106, 153, 85, 0.5);
							}
						}
					}
				}
			}
		}
	}
}
</style>

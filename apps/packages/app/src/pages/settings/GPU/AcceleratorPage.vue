<template>
	<page-title-component :title="t('home_menus.gpu')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div v-if="store.nodes.length > 0">
			<app-menu-feature :menu-type="MENU_TYPE.GPU" />

			<div
				v-for="node in store.nodes"
				:key="node.nodeName"
				class="accelerator-node-card q-mt-md"
			>
				<!-- node header -->
				<div
					class="accelerator-node-header row justify-between items-center q-pa-lg no-wrap"
					style="gap: 8px"
				>
					<div class="row items-center no-wrap" style="min-width: 0">
						<div class="accelerator-node-icon row items-center justify-center">
							<img
								v-if="nodeIconAsset(node)"
								:src="nodeIconAsset(node) as string"
								class="accelerator-node-icon__img"
							/>
							<q-icon v-else name="sym_r_dns" size="24px" class="text-ink-2" />
						</div>
						<div class="column q-ml-md" style="min-width: 0">
							<div class="row items-center no-wrap" style="gap: 8px">
								<div class="text-h6 text-ink-1 ellipsis">
									{{ node.nodeName }}
								</div>
								<div class="row items-center" style="gap: 4px">
									<compute-resource-badge
										v-for="badge in nodeBadges(node)"
										:key="badge"
										:label="badge"
										tone="neutral"
									/>
								</div>
							</div>
							<div
								class="row items-center q-mt-xs accelerator-node-summary"
								style="gap: 12px"
							>
								<span
									v-for="tile in summaryTiles(node)"
									:key="tile.label"
									class="row items-center no-wrap text-body3"
									style="gap: 4px"
								>
									<span class="text-ink-2">{{ tile.label }}: </span>
									<span class="text-blue-default">{{ tile.value }}</span>
									<q-icon
										v-if="tile.hint"
										size="16px"
										name="sym_r_info"
										class="text-ink-3"
									>
										<q-tooltip
											self="top left"
											class="text-body3"
											:offset="[0, 0]"
										>
											{{ tile.hint }}
										</q-tooltip>
									</q-icon>
								</span>
							</div>
						</div>
					</div>
					<q-btn
						flat
						dense
						no-caps
						class="accelerator-manage-btn text-body3 text-ink-2 q-px-md"
						:label="t('Manage')"
						@click="goManage(node.nodeName)"
					/>
				</div>

				<!-- devices -->
				<div class="accelerator-node-body column q-pa-lg" style="gap: 12px">
					<accelerator-device-card
						v-for="device in node.devices || []"
						:key="device.id"
						:device="device"
						:apps="store.deviceApps(device)"
					/>
				</div>
			</div>
		</div>

		<div
			v-else-if="store.loading"
			class="accelerator-loading column items-center justify-center"
		>
			<bt-loading :loading="true" size="48px" />
			<div
				class="text-body2 text-ink-2 q-mt-md text-center"
				style="white-space: pre-line"
			>
				{{ t('Getting accelerator status.\nPlease wait...') }}
			</div>
		</div>

		<app-menu-empty
			v-else
			:menu-type="MENU_TYPE.GPU"
			:title="t('No accelerator found')"
			:message="
				t('Connect a GPU or other accelerator to your cluster to get started.')
			"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import BtLoading from 'src/components/base/BtLoading.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import ComputeResourceBadge from 'src/components/compute/ComputeResourceBadge.vue';
import AcceleratorDeviceCard from 'src/components/compute/AcceleratorDeviceCard.vue';
import nodeOlaresOneIcon from 'src/assets/accelerator/node-olares-one.png';
import nodeZeroIcon from 'src/assets/accelerator/node-zero.png';
import { MENU_TYPE } from 'src/constant';
import {
	COMPUTE_TYPE,
	ComputeResourceNode,
	computeModeTitle,
	formatComputeMemory,
	isVramComputeMode,
	normalizeComputeMode
} from 'src/constant/compute';
import { useAcceleratorStore } from 'src/stores/settings/accelerator';
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const router = useRouter();
const store = useAcceleratorStore();

onMounted(() => {
	store.fetch();
});

function nodeBadges(node: ComputeResourceNode): string[] {
	return (node.gpuTypes ?? []).map((mode) => computeModeTitle(String(mode)));
}

// Best-effort node illustration; returns null to fall back to the generic icon.
function nodeIconAsset(node: ComputeResourceNode): string | null {
	const name = `${node.nodeName ?? ''}`.toLowerCase();
	// i18n-ignore-start: node name matching for icon asset, not user-visible
	if (name.includes('olares one') || name.includes('olares-one')) {
		return nodeOlaresOneIcon;
	}
	if (name.includes('olares zero') || name.includes('olares-zero')) {
		return nodeZeroIcon;
	}
	// i18n-ignore-end
	return null;
}

interface SummaryTile {
	label: string;
	value: string;
	hint?: string;
}

function summaryTiles(node: ComputeResourceNode): SummaryTile[] {
	let vramTotal = 0;
	let sharedTotal = 0;

	(node.devices || []).forEach((device) => {
		const total = device.memory || 0;
		// CPU has no VRAM/memory capacity to report here; skip it entirely.
		if (normalizeComputeMode(device.mode) === COMPUTE_TYPE.CPU) {
			return;
		}
		if (isVramComputeMode(device.mode)) {
			vramTotal += total;
		} else {
			sharedTotal += total;
		}
	});

	const tiles: SummaryTile[] = [];
	if (vramTotal > 0) {
		tiles.push({
			label: t('Dedicated VRAM'),
			value: formatComputeMemory(vramTotal)
		});
	}
	if (sharedTotal > 0) {
		tiles.push({
			label: t('Shared RAM'),
			value: formatComputeMemory(sharedTotal),
			hint: t('Shared RAM is set to 75% of your total system RAM')
		});
	}
	return tiles;
}

function goManage(nodeName: string) {
	router.push(`/gpu/manage/${encodeURIComponent(nodeName)}`);
}
</script>

<style scoped lang="scss">
.accelerator-node-card {
	border: 1px solid $separator;
	border-radius: 12px;
	background-color: $background-1;
}

.accelerator-node-header {
	border-bottom: 1px solid $separator;
}

.accelerator-node-icon {
	width: 40px;
	height: 40px;
	border-radius: 8px;
	background-color: $background-3;
	overflow: hidden;
	flex-shrink: 0;

	&__img {
		width: 28px;
		height: 28px;
		object-fit: contain;
	}
}

.accelerator-manage-btn {
	border: 1px solid $btn-stroke;
	border-radius: 8px;
	flex-shrink: 0;
	height: 34px !important;
}

.accelerator-node-summary {
	flex-wrap: wrap;
}

.accelerator-loading {
	min-height: 240px;
}
</style>

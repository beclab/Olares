<template>
	<div class="ports-wrapper q-gutter-y-md" v-if="ports">
		<div v-for="(port, index) in ports" :key="index" class="ports q-pa-lg">
			<div class="row ports-item items-center no-wrap q-col-gutter-x-md">
				<ListItem
					size="sm"
					icon-size="sm"
					:title-style="{ marginRight: '14px' }"
					:title="port.targetPort"
					:sub-title="t('CONTAINER_PORT_SCAP')"
				>
					<template #avatar>
						<q-img :src="servicesIcon" />
					</template>
				</ListItem>
				<div class="text-grey-5 q-mx-lg protocol-style">
					{{ port.protocol }}
				</div>
				<ListItem
					size="sm"
					icon-size="sm"
					:title-style="{ marginRight: '14px' }"
					:title="port.port"
					:sub-title="t('SERVICE_PORT_SCAP')"
				>
					<template #avatar>
						<q-img :src="ringIcon" />
					</template>
				</ListItem>
				<template v-if="port.nodePort">
					<div class="text-grey-5 q-mx-lg protocol-style">
						{{ port.protocol }}
					</div>
					<ListItem
						size="sm"
						icon-size="sm"
						:title-style="{ marginRight: '14px' }"
						:title="port.nodePort"
					>
						<template #avatar>
							<q-img :src="indication2Icon" />
						</template>
						<template #subTitle>
							{{ t('NODE_PORT_SCAP') }}
							<q-icon name="sym_r_info" color="grey-5">
								<q-tooltip>
									<div style="max-width: 240px">
										{{ t('SERVICE_NODE_PORT_DESC') }}
									</div>
								</q-tooltip>
							</q-icon>
						</template>
					</ListItem>
				</template>
			</div>
		</div>
	</div>
	<Empty v-if="noData"></Empty>
</template>

<script lang="ts" setup>
import { defineProps, computed } from 'vue';
import { isEmpty } from 'lodash';
import servicesIcon from '../assets/server-blue.svg';
import indication2Icon from '../assets/menu-button-wide.svg';
import ringIcon from '../assets/topo-node-nslice.svg';
import ListItem from '../components/MyListItem/ListItem.vue';
import { useI18n } from 'vue-i18n';
import Empty from '../components/Empty.vue';
const { t } = useI18n();
interface Props {
	ports: any;
}
const props = withDefaults(defineProps<Props>(), {});
const noData = computed(() => isEmpty(props.ports));
</script>

<style scoped lang="scss">
.ports-wrapper {
	white-space: nowrap;
	.ports {
		position: relative;
		border-radius: 8px;
		border: 1px solid $separator;
		.protocol-style {
			font-size: 10px;
		}
	}
}
</style>

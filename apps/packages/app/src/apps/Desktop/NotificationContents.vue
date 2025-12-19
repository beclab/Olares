<template>
	<div v-for="(item, index) in notificationStore.data" :key="item.id">
		<div class="title" style="margin-top: 32px" v-if="item.open">
			<div class="brige">{{ item.appName }}</div>
			<div class="">
				<span
					class="cancel q-ml-sm"
					@click="notificationStore.deleteItem(item, -1)"
				>
					<span class="clearTxt">Clear All</span>
					<span class="cancel-icon">
						<q-icon class="icon" name="close" size="12px" />
					</span>
				</span>

				<span class="less" @click="item.open = false">Show less</span>
				<span class="less-icon">
					<q-icon class="icon" name="expand_less" size="12px" />
				</span>
			</div>
		</div>

		<div
			class="content"
			style="margin-top: 30px"
			v-if="item.childrens.length > 1 && !item.open"
		>
			<div class="notify-item" @click="item.open = true">
				<div class="avatar" v-if="item.icon">
					<q-img :src="item.icon" style="width: 36px; height: 36px" />
				</div>
				<div class="info">
					<div class="row items-center justify-between">
						<div class="tit text-subtitle2 text-ink1">
							{{ item.childrens.slice(-1)[0].title }}
						</div>
						<div class="q-ml-lg text-body3 text-ink-2">
							{{ formatStampTime(item.childrens.slice(-1)[0].createTime) }}
						</div>
					</div>
					<div class="message text-body3 text-ink1">
						{{ item.childrens.slice(-1)[0].body }}
					</div>
				</div>

				<span class="removeApp" @click="notificationStore.deleteItem(item, -1)">
					<q-icon class="icon" name="close" size="12px" />
				</span>
			</div>
			<div class="multiple-item"></div>
			<div class="multiple-item2"></div>
			<!-- </template> -->
		</div>

		<div class="content" v-else>
			<div
				class="notify-item"
				v-for="(cell, cellIndex) in item.childrens"
				:key="cellIndex"
			>
				<div class="avatar" v-if="item.icon">
					<q-img :src="item.icon" style="width: 36px; height: 36px" />
				</div>
				<div class="info">
					<div class="row items-center justify-between">
						<div class="tit text-subtitle2 text-ink1">
							{{ cell.title }}
						</div>
						<div class="q-ml-lg text-body3 text-ink-2">
							{{ formatStampTime(cell.createTime) }}
						</div>
					</div>
					<div class="message text-body3 text-ink1">{{ cell.body }}</div>
				</div>

				<span
					class="removeApp"
					@click="notificationStore.deleteItem(item, cellIndex)"
				>
					<q-icon class="icon" name="close" size="12px" />
				</span>
			</div>
		</div>
	</div>
	<slot name="more" />
</template>

<script setup lang="ts">
import { useNotificationStore } from '../../stores/desktop/notification';
import { date } from 'quasar';
const notificationStore = useNotificationStore();

const formatStampTime = (createTime: number) => {
	const currentDate = new Date();
	const createDate = new Date(createTime);
	const daysDifference = Math.abs(currentDate.getDay() - createDate.getDay());
	const monthsDifference = Math.abs(
		currentDate.getMonth() - createDate.getMonth()
	);
	const yearsDifference = Math.abs(
		currentDate.getFullYear() - createDate.getFullYear()
	);

	if (yearsDifference > 0) {
		return date.formatDate(createTime, 'YYYY-MM-DD');
	} else if (monthsDifference > 0) {
		return date.formatDate(createTime, 'MM-DD');
	} else if (daysDifference > 0) {
		return date.formatDate(createTime, 'MM-DD HH:mm');
	}

	return date.formatDate(createTime, 'HH:mm:ss');
};
</script>

<style scoped lang="scss">
.title {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 12px;
	.brige {
		color: #ffffff;
		font-weight: 700;
		font-size: 16px;
	}
	.less {
		display: inline-block;
		font-size: 12px;
		height: 22px;
		line-height: 22px;
		color: #5c5551;
		padding: 0 12px;
		border-radius: 12px;
		background: rgba(246, 246, 246, 0.4);
		box-shadow: 0px 0px 40px 0px rgba(0, 0, 0, 0.2),
			0px 0px 2px 0px rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(30px);
		cursor: pointer;
		float: right;
	}
	.less-icon {
		width: 22px;
		height: 22px;
		border-radius: 11px;
		background: rgba(246, 246, 246, 0.4);
		box-shadow: 0px 0px 40px 0px rgba(0, 0, 0, 0.2),
			0px 0px 2px 0px rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(30px);
		display: flex;
		align-items: center;
		justify-content: center;
		float: right;
		opacity: 0;
		.icon {
			color: #5c5551;
		}
	}
	.cancel {
		display: inline-block;
		height: 22px;
		line-height: 22px;
		font-size: 12px;
		color: #5c5551;
		padding: 0px 11px;
		border-radius: 11px;
		background: rgba(246, 246, 246, 0.4);
		box-shadow: 0px 0px 40px 0px rgba(0, 0, 0, 0.2),
			0px 0px 2px 0px rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(30px);
		cursor: pointer;
		position: relative;
		float: right;
		.cancel-icon {
			width: 22px;
			height: 22px;
			line-height: 21px;
			text-align: center;
			display: inline-block;
			color: rgba(246, 246, 246, 0.4);
			position: absolute;
			right: 0px;
			top: 0px;
			.icon {
				color: #5c5551;
			}
		}
		.clearTxt {
			display: none;
			float: right;
			transition: all 1s ease-in-out;
		}
		&:hover {
			.clearTxt {
				display: inline-block;
			}
			.cancel-icon {
				display: none;
			}
		}
		&:hover + .less {
			display: none;
		}
		&:hover ~ .less-icon {
			opacity: 1;
		}
	}
}

.content {
	position: relative;

	.notify-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		fill: rgba(246, 246, 246, 0.5);
		stroke-width: 1px;
		stroke: rgba(255, 255, 255, 0.2);
		backdrop-filter: blur(120px);
		background: linear-gradient(
				0deg,
				rgba(246, 246, 246, 0.5),
				rgba(246, 246, 246, 0.5)
			),
			linear-gradient(0deg, rgba(255, 255, 255, 0.2), rgba(255, 255, 255, 0.2));

		border: 1px solid #ffffff33;

		padding: 12px 20px;
		border-radius: 20px;
		margin-bottom: 8px;

		.avatar {
			width: 36px;
			height: 36px;
			border-radius: 8px;
			overflow: hidden;
			margin-right: 12px;
		}
		.info {
			flex: 1;
			width: calc(100% - 50px);
			.tit {
				flex: 1;
				overflow: hidden;
				line-height: 20px;
				text-overflow: ellipsis;
				white-space: nowrap;
			}
		}

		.removeApp {
			position: absolute;
			top: -4px;
			left: -4px;
			display: none;
			width: 20px;
			height: 20px;
			line-height: 19px;
			text-align: center;
			border-radius: 11px;
			fill: rgba(246, 246, 246, 0.5);
			stroke-width: 1px;
			stroke: rgba(255, 255, 255, 0.2);
			backdrop-filter: blur(60px);
			background: linear-gradient(
					0deg,
					rgba(246, 246, 246, 0.5),
					rgba(246, 246, 246, 0.5)
				),
				linear-gradient(
					0deg,
					rgba(255, 255, 255, 0.2),
					rgba(255, 255, 255, 0.2)
				);
			position: absolute;
			right: -1px;
			top: -1px;
			cursor: pointer;
			.icon {
				color: #5c5551;
			}
			&:hover {
				background: rgba(230, 230, 230, 1);
			}
		}

		&:hover > .removeApp {
			display: inline-block;
		}
	}
	.multiple-item {
		width: 320px;
		height: 30px;
		fill: rgba(246, 246, 246, 0.4);
		stroke-width: 1px;
		stroke: rgba(255, 255, 255, 0.2);
		border-radius: 12px;
		border: 1px solid #ffffff33;
		overflow: hidden;
		backdrop-filter: blur(60px);
		position: absolute;
		bottom: -8px;
		left: 0;
		right: 0;
		margin: auto;
		background: rgba(198, 198, 198, 0.3);
		z-index: -1;
		&:hover > .removeApp {
			display: inline-block;
		}
	}
	.multiple-item2 {
		width: 300px;
		height: 30px;
		fill: rgba(246, 246, 246, 0.4);
		stroke-width: 1px;
		stroke: rgba(255, 255, 255, 0.2);
		border-radius: 12px;
		border: 1px solid #ffffff33;
		overflow: hidden;
		backdrop-filter: blur(60px);
		position: absolute;
		bottom: -16px;
		left: 0;
		right: 0;
		margin: auto;
		background: rgba(198, 198, 198, 0.3);
		z-index: -2;
		&:hover > .removeApp {
			display: inline-block;
		}
	}
}
</style>

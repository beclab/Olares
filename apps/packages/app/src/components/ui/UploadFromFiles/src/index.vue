<template>
	<div class="UploadFromFiles" v-show="show">
		<div class="title">Upload File</div>
		<div class="content" ref="FilesRef">
			<ul v-for="item in fileData" :key="item.path">
				<li
					:class="it.isActive ? 'isActive' : ''"
					v-for="it in item.items"
					@click="handleClick(it)"
					:key="it.path"
				>
					{{ it.name }}
				</li>
			</ul>
			<ul v-if="fileDesc">
				<li>
					<span>Display Name:</span>
					<span>{{ fileDesc.name }}</span>
				</li>
				<li>
					<span>Size:</span>
					<span>{{ sizeFormatting(fileDesc.size) }}</span>
				</li>
				<li>
					<span>Type:</span>
					<span>{{ fileDesc.type }}</span>
				</li>
				<li>
					<span>Last Modified:</span>
					<span>{{
						moment(fileDesc.modified).format('MMMM Do YYYY, h:mm:ss a')
					}}</span>
				</li>
				<li>
					<span>Path:</span>
					<span>{{ fileDesc.path }}</span>
				</li>
			</ul>
		</div>

		<div class="footer">
			<span class="cancel" @click="cancel">Cancel</span>
			<span class="confirm" v-if="fileDesc" @click="confirm">Confirm</span>
			<span class="confirmDisabled" v-else>Confirm</span>
		</div>
	</div>
	<div class="mask" v-show="show"></div>
</template>

<script lang="ts">
import { defineComponent, ref, watch, onMounted } from 'vue';
import axios from 'axios';
import moment from 'moment';

export default defineComponent({
	name: 'UploadFromFiles',
	props: {
		show: {
			type: Boolean,
			default: false
		},
		url: {
			type: String,
			default: ''
		}
	},
	setup(props, ctx) {
		const fileDesc = ref();
		const fileData = ref<any[]>([]);

		const handleClick = (item: any) => {
			for (let i = 0; i < fileData.value.length; i++) {
				const el1 = fileData.value[i];
				if (el1.items && el1.items.length > 0) {
					for (let j = 0; j < el1.items.length; j++) {
						const el2 = el1.items[j];
						const clickIndex = el1.items.findIndex(
							(cell: { path: any }) => cell.path === item.path
						);
						if (clickIndex >= 0) {
							el2.isActive = false;
							if (el2.path === item.path) {
								el2.isActive = true;
							}
						}
					}
				}
			}

			fileDesc.value = null;
			if (!item.isDir) {
				fileDesc.value = item;
			}
			chooseFile(item.path, item.isDir);
		};

		const chooseFile = async (path: string, isDir: boolean) => {
			const tempArr = new Array<any>();
			for (let i = 0; i < fileData.value.length; i++) {
				const el1 = fileData.value[i];
				let breakFlag = false;
				if (el1.items && el1.items.length >= 0) {
					for (let j = 0; j < el1.items.length; j++) {
						const el2 = el1.items[j];
						if (el2.path === path) {
							breakFlag = true;
						}
					}
				}

				tempArr.push(el1);
				if (breakFlag) break;
			}

			if (isDir) {
				const data = await getData(path);
				tempArr.push(data);
			}
			fileData.value = tempArr;
		};

		const sizeFormatting = (now: number) => {
			if (now === 0) {
				return now;
			}
			const unitArr = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
			let index = 0,
				srcsize = now;
			index = Math.floor(Math.log(srcsize) / Math.log(1024));
			let size = srcsize / Math.pow(1024, index);
			if (index !== 0) {
				size = Number(size.toFixed(2));
			}
			return size + ' ' + unitArr[index];
		};

		const cancel = () => {
			ctx.emit('cancel');
		};

		const confirm = () => {
			ctx.emit('confirm', fileDesc.value.path);
		};

		const getData = async (path: string) => {
			const URL = props.url;
			const instance = axios.create({
				baseURL: URL,
				timeout: 2000
			});

			return await instance
				.get(`/api/resources${path}`, {})
				.then(function (res) {
					return res.data;
				})
				.catch(function (err) {
					console.log('err', err);
				});
		};

		watch(
			() => props.show,
			async (newVal) => {
				if (newVal) {
					const data = await getData('/Home');
					fileData.value.push(data);
				} else {
					fileData.value = [];
					fileDesc.value = null;
				}
			}
		);

		return {
			fileData,
			fileDesc,
			moment,
			cancel,
			confirm,
			handleClick,
			sizeFormatting
		};
	}
});
</script>

<style lang="scss" scoped>
ul,
li {
	margin: 0;
	padding: 0;
	list-style: none;
}
.mask {
	width: 100%;
	height: 100%;
	background-color: rgba(0, 0, 0, 0.25);
	position: fixed;
	top: 0;
	left: 0;
	z-index: 9999;
}
.UploadFromFiles {
	width: 800px;
	height: 400px;
	position: fixed;
	left: 0;
	top: 0;
	right: 0;
	bottom: 0;
	margin: auto;
	border-radius: 10px;
	overflow: hidden;
	text-align: left;
	background-color: #ffffff;
	z-index: 10000;

	.title {
		width: 100%;
		height: 40px;
		line-height: 40px;
		background-color: #ffffff;
		border-bottom: 1px solid #e5e5e5;
		text-align: left;
		color: #000000;
		text-indent: 20px;
		font-size: 14px;
		font-weight: 600;
	}

	.content {
		height: calc(100% - 92px);
		overflow-x: scroll;
		white-space: nowrap;
		border-bottom: 1px solid #e5e5e5;
		ul {
			width: 30%;
			height: calc(100% - 8px);
			border-right: 1px solid #e5e5e5;
			overflow-y: scroll;
			display: inline-block;
			padding-top: 8px;
			> li {
				cursor: pointer;
				padding: 0 8px;
				margin-left: 8px;
				margin-right: 8px;
				border-radius: 4px;
				overflow: hidden;
				font-size: 14px;
				line-height: 26px;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
				&.isActive {
					background-color: rgba(25, 118, 210, 0.9);
					color: #ffffff;
				}
			}

			&:last-child {
				width: 40%;
				> li {
					display: flex;
					align-items: flex-start;
					justify-content: space-between;
					white-space: normal;
					cursor: auto;
					span:first-child {
						width: 100px;
					}
					span:last-child {
						float: right;
						flex: 1;
						padding-left: 10px;
					}
				}
			}
		}
	}

	.footer {
		width: 100%;
		height: 50px;
		line-height: 50px;
		display: flex;
		justify-content: flex-end;
		align-items: center;
		span {
			display: inline-block;
			height: 34px;
			line-height: 34px;
			text-align: center;
			border-radius: 4px;
			font-size: 14px;
			padding: 0 10px;
			margin-right: 20px;
			cursor: pointer;
			&:hover {
				opacity: 0.8;
			}
		}

		.cancel {
			background-color: #e5e5e5;
			color: #3f3f3f;
		}

		.confirm {
			background-color: rgba(25, 118, 210, 0.9);
			color: #ffffff;
		}
		.confirmDisabled {
			background-color: #e5e5e5;
			color: #3f3f3f;
		}
	}
}
</style>

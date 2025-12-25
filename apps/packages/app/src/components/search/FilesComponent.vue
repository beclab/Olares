<template>
	<div class="searchCard">
		<div class="appIcon">
			<img class="icon" :src="'./app-icon/files.svg'" alt="search" />
			<span>Files</span>
			<img
				src="../../assets/desktop/delete-icon.svg"
				alt="delete"
				@click="goBack"
			/>
		</div>
		<input
			class="input"
			type="text"
			ref="filesRef"
			v-model.trim="searchFiles"
			placeholder="Search files..."
		/>
	</div>
	<div class="files">
		<div class="fileList">
			<template v-for="(item, index) in fileData" :key="item.name">
				<div
					ref="isActiveRef"
					:class="[activeItem === item.name ? 'isActive' : '', 'txt']"
					@click="chooseFile(index)"
				>
					<img
						:src="
							item?.fileType
								? `./files/${item?.fileType}.svg`
								: `./files/other.svg`
						"
					/>
					<span>{{ item.name }}</span>
				</div>
			</template>
		</div>
		<div class="fileDetails" v-if="fileData && fileData.length > 0">
			<div class="header">
				<div class="desc">
					<img
						class="arrowLeft"
						src="../../assets/desktop/arrowLeft.svg"
						alt="search"
					/>
					<span class="name">{{ fileDetail?.name }}</span>
				</div>
				<div
					v-if="
						fileDetail &&
						['.doc', '.docx', '.pdf', '.txt', '.md'].includes(fileDetail.type)
					"
				>
					<span class="actions" @click="handleAction('summarize', fileDetail)"
						>Summary</span
					>
					<span class="actions" @click="handleAction('askAi', fileDetail)"
						>Ask AI</span
					>
				</div>
			</div>
			<div class="logo">
				<img
					:src="
						fileDetail?.fileType
							? `./files/${fileDetail?.fileType}.svg`
							: `./files/other.svg`
					"
				/>
			</div>
			<div class="li">
				<span>Name</span><span>{{ fileDetail?.name }}</span>
			</div>
			<div class="li">
				<span>Type</span><span>{{ fileDetail?.type }}</span>
			</div>
			<div class="li">
				<span>Size</span><span>{{ fileDetail?.size }}</span>
			</div>
			<div class="li">
				<span>Created</span><span>{{ fileDetail?.created }}</span>
			</div>
			<div class="li">
				<span>Modified</span><span>{{ fileDetail?.modified }}</span>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { debounce } from 'quasar';
import { useSearchStore } from 'src/stores/search';

interface TerminusParams {
	from: string;
	message: string;
	path: string;
	fileType: string;
}

export default {
	name: 'FilesComponent',
	props: {
		showSearchDialog: {
			type: Boolean,
			require: false,
			default: false
		},
		handSearchFiles: {
			type: String,
			require: false
		}
	},
	components: {},
	defineProps: [],

	setup(props: any, context: any) {
		const searchFiles = ref('');
		const filesRef = ref();

		const fileDetail = ref();
		const fileData = ref();
		const activeItem = ref();
		const isActiveRef = ref();
		const searchStore = useSearchStore();
		const pattern = new RegExp('[\u4E00-\u9FA5]+');

		const chooseFile = (index: number) => {
			fileDetail.value = fileData.value[index];
			activeItem.value = fileData.value[index].name;
		};

		const classifyData = (data: any) => {
			const newArr = data;
			activeItem.value = newArr[0]?.name;
			fileDetail.value = newArr[0];
			return newArr;
		};

		const refreshScroll = (index: number) => {
			isActiveRef.value[index].scrollIntoView({
				behavior: 'auto',
				block: 'nearest'
			});
		};

		watch(
			() => searchFiles.value,
			debounce(async (newVal: string | undefined) => {
				if (!newVal) {
					return (fileData.value = []);
				}
				if (!pattern.test(newVal) && newVal.length <= 2) {
					return false;
				}

				const filesList = await getFiles(newVal);
				fileData.value = await classifyData(filesList);
			}, 600)
		);

		const getFiles = (query?: string) => {
			return searchStore.getFiles(query);
		};

		const handleAction = (from: string, item: any) => {
			let terminusParams: TerminusParams | null = null;
			let message = '';
			if (from === 'summarize') {
				message = 'Please help me summarize this document';
			}
			terminusParams = {
				from,
				message,
				path: item.where,
				fileType: item.fileType
			};

			context.emit('handleAction', terminusParams);
		};

		const goBack = () => {
			context.emit('goBack');
		};

		const keydownEnter = (event: any) => {
			if (event.keyCode === 8 && !searchFiles.value) {
				return goBack();
			}

			if (!fileData.value) return false;
			const index = fileData.value.findIndex(
				(item: { name: any }) => item.name === activeItem.value
			);
			const upIndex = index - 1;
			const downIndex = index + 1;

			switch (event.keyCode) {
				case 38:
					if (upIndex >= 0) {
						activeItem.value = fileData.value[upIndex].name;
						refreshScroll(upIndex);
						chooseFile(upIndex);
					}
					break;

				case 40:
					if (downIndex <= fileData.value.length - 1) {
						activeItem.value = fileData.value[downIndex].name;
						refreshScroll(downIndex);
						chooseFile(downIndex);
					}
					break;
			}
		};

		onMounted(async () => {
			filesRef.value && filesRef.value.focus();
			if (props.handSearchFiles) searchFiles.value = props.handSearchFiles;
			window.addEventListener('keydown', keydownEnter);
		});

		onUnmounted(() => {
			window.removeEventListener('keydown', keydownEnter);
		});

		return {
			fileDetail,
			fileData,
			activeItem,
			isActiveRef,
			searchFiles,
			filesRef,
			chooseFile,
			handleAction,
			goBack
		};
	}
};
</script>

<style lang="scss" scoped>
.files {
	height: calc(100% - 76px);
	border-top: 1px solid #d8d8d8;
	overflow: scroll;
	display: flex;

	.fileList {
		width: 224px;
		padding-top: 8px;
		border-right: 1px solid #d8d8d8;
		overflow: scroll;

		.txt {
			width: 216px;
			height: 36px;
			line-height: 36px;
			display: flex;
			align-items: center;
			border-radius: 4px;
			cursor: pointer;
			margin: 0 auto;
			img {
				width: 24px;
				height: 24px;
				border: 0.6px solid #e0e0e0;
				border-radius: 4px;
				margin: 0 8px;
			}
			span {
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
			}
			&:hover {
				background: rgba(26, 19, 15, 0.06);
			}
			&.isActive {
				background: rgba(26, 19, 15, 0.06);
			}
		}
	}
	.fileDetails {
		width: calc(100% - 224px);
		overflow: scroll;
		.header {
			height: 44px;
			line-height: 44px;
			border-bottom: 1px solid #d8d8d8;
			display: flex;
			align-items: center;
			justify-content: space-between;
			.desc {
				display: flex;
				align-items: center;
				color: #857c77;
				.arrowLeft {
					width: 7px;
					height: 13px;
					margin: 0 12px 0 20px;
				}
				.name {
					max-width: 340px;
					display: inline-block;
					overflow: hidden;
					white-space: nowrap;
					text-overflow: ellipsis;
				}
			}

			.actions {
				height: 32px;
				padding: 0 8px;
				display: inline-block;
				line-height: 32px;
				text-align: center;
				background: rgba(26, 19, 15, 0.06);
				border-radius: 4px;
				margin-right: 16px;
				cursor: pointer;
				&:hover {
					background: rgba(26, 19, 15, 0.1);
				}
			}
		}
		.logo {
			text-align: center;
			margin: 20px 0;
			img {
				width: 100px;
				height: 100px;
			}
		}
		.li {
			height: 38px;
			line-height: 38px;
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding: 0 16px;
			border-bottom: 1px solid #d8d8d8;
			color: #857c77;
		}
	}
}
</style>

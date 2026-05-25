export const importFilesStyle = (isMobile = false) => {
	if (isMobile) {
		import('../../css/common/files/listing-mobile.scss' as any).then(() => {});
	}
	import('../../css/common/files/listing.scss' as any).then(() => {});
};

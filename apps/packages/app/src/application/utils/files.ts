export const importFilesStyle = (isMobile = false) => {
	if (isMobile) {
		import('../../css/listing-mobile.css' as any).then(() => {});
	}
	import('../../css/listing.scss' as any).then(() => {});
};

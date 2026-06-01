/**
 * Frontend rules aligned with image upload API (png / jpeg / jpg / gif).
 */

const UPLOAD_V1_EXT_ORDERED = ['jpg', 'jpeg', 'png', 'gif'] as const;

/** HTML `accept` for file inputs aligned with `/images/upload/v1`. */
export const IMAGES_UPLOAD_V1_ACCEPT = UPLOAD_V1_EXT_ORDERED.map(
	(ext) => `.${ext}`
).join(',');

const ALLOWED_EXT = new Set<string>(UPLOAD_V1_EXT_ORDERED);
const ALLOWED_MIME = new Set(['image/png', 'image/jpeg', 'image/gif']);

export function matchesImagesUploadV1Formats(file: File): boolean {
	const name = file.name || '';
	const ext = name.includes('.')
		? name.slice(name.lastIndexOf('.') + 1).toLowerCase()
		: '';
	const mime = (file.type || '').toLowerCase();

	if (mime && ALLOWED_MIME.has(mime)) {
		return true;
	}
	if (mime && mime !== 'application/octet-stream') {
		return false;
	}
	return ALLOWED_EXT.has(ext);
}

export function createImagesUploadV1FormatGuard(
	t: (key: string) => string
): (file: File) => string | null {
	return (file: File) =>
		matchesImagesUploadV1Formats(file)
			? null
			: t('bt_uploader_image_format_not_allowed');
}

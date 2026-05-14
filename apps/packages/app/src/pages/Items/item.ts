import { app } from '../../globals';
export interface TagInter {
	name: string;
	icon: string;
	class: string;
}

export const showTags = (item: any) => {
	const tags: TagInter[] = [];
	let tagWidth = 0;
	if (item.tags.length) {
		tags.push({
			icon: 'sym_r_style',
			name: item.tags[0],
			class: ''
		});
		tagWidth += 100;
		if (item.tags.length > 1) {
			tags.push({
				icon: 'sym_r_style',
				name: `+${item.tags.length - 1}`,
				class: ''
			});
			tagWidth += 54;
		}
	}
	const attCount = (item.attachments && item.attachments.length) || 0;
	if (attCount) {
		tags.push({
			name: attCount.toString(),
			icon: 'sym_r_attach_file',
			class: ''
		});
		tagWidth += 42;
	}
	if (app.account!.favorites.has(item.id)) {
		tags.push({
			name: '',
			icon: 'sym_r_grade',
			class: 'text-red'
		});
		tagWidth += 32;
	}

	return {
		tags,
		tagWidth
	};
};

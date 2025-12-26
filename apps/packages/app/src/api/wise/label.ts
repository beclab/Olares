import axios, { CancelToken } from 'axios';
import { useConfigStore } from 'src/stores/rss-config';
import { Label } from 'src/utils/rss-types';

export async function syncLabels(time: number): Promise<Label[]> {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/label/sync/' + time);
}

export async function createLabel(name: string): Promise<Label> {
	const configStore = useConfigStore();
	return await axios.post(configStore.url + '/knowledge/label', {
		name
	});
}

export async function updateLabel(label: Label): Promise<Label> {
	const configStore = useConfigStore();
	return await axios.put(
		configStore.url + '/knowledge/label/' + label.id,
		label
	);
}

export async function deleteLabel(id: string): Promise<Label> {
	const configStore = useConfigStore();
	return await axios.delete(configStore.url + '/knowledge/label/' + id);
}

export async function setLabelOnEntry(
	label_id: string,
	entry_id: string
): Promise<Label> {
	const configStore = useConfigStore();
	return await axios.post(
		configStore.url + '/knowledge/label/entry/' + label_id,
		{
			entry_id
		}
	);
}

export async function removeLabelOnEntry(
	label_id: string,
	entry_id: string
): Promise<Label> {
	const configStore = useConfigStore();
	return await axios.delete(
		configStore.url + '/knowledge/label/entry/' + label_id,
		{
			data: {
				entry_id
			}
		}
	);
}

export async function getLabel(
	entry_id: string,
	cancelToken?: CancelToken
): Promise<{ items: Label[] }> {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/label', {
		params: {
			entry_id
		},
		cancelToken
	});
}

export async function setLabelOnNote(
	label_id: string,
	note_id: string
): Promise<Label> {
	const configStore = useConfigStore();
	return await axios.post(
		configStore.url + '/knowledge/label/note/' + label_id,
		{
			note_id
		}
	);
}

export async function removeLabelOnNote(
	label_id: string,
	note_id: string
): Promise<Label> {
	const configStore = useConfigStore();
	return await axios.delete(
		configStore.url + '/knowledge/label/note/' + label_id,
		{
			data: {
				note_id
			}
		}
	);
}

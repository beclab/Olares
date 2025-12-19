import axios from 'axios';
import { useConfigStore } from 'src/stores/rss-config';
import { CreateNote, Note } from 'src/utils/rss-types';

export async function syncNotes(time: number): Promise<Note[]> {
	const configStore = useConfigStore();
	return await axios.get(configStore.url + '/knowledge/note/sync/' + time);
}

export async function createNote(note: CreateNote): Promise<Note> {
	const configStore = useConfigStore();
	return await axios.post(configStore.url + '/knowledge/note', note);
}

export async function updateNote(note: Note): Promise<Note> {
	const configStore = useConfigStore();
	return await axios.put(configStore.url + '/knowledge/note/' + note.id, note);
}

export async function deleteNote(id: string): Promise<Note> {
	const configStore = useConfigStore();
	return await axios.delete(configStore.url + '/knowledge/note/' + id);
}

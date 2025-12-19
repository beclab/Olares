import { getWordList } from './wordlists';
import { randomNumber } from '@didvault/sdk/src/core/util';

export async function generatePassphrase(
	nWords = 4,
	separator = '-',
	languages = ['en']
) {
	const words: any = [];
	const list: string[] = [];

	for (const lang of languages) {
		list.push(...(await getWordList(lang)));
	}

	if (!list.length) {
		list.push(...(await getWordList('en')));
	}

	for (let i = 0; i < nWords; i++) {
		words.push(list[await randomNumber(0, list.length - 1)]);
	}

	return words.join(separator);
}

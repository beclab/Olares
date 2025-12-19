import axios from 'axios';
import { DomainCookie } from 'src/constant/constants';

export async function deleteCookie(
	base_url: string,
	key: string
): Promise<void> {
	return await axios.post(`${base_url}/api/cookie/delete`, { key });
}

export async function getCookies(base_url: string): Promise<DomainCookie[]> {
	return await axios.get(`${base_url}/api/cookie/all`);
}

export async function createOrUpdateCookie(
	base_url: string,
	domainCookie: DomainCookie
) {
	return await axios.post(`${base_url}/api/cookie`, domainCookie);
}

export async function createCookieList(
	base_url: string,
	domainCookie: DomainCookie[]
) {
	return await axios.post(`${base_url}/api/cookie/list`, domainCookie);
}

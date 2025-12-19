const hp = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
export function generatePasword() {
	let password = '';
	for (let i = 0; i < 16; ++i) {
		const index = Math.floor(Math.random() * hp.length);
		password = password + hp[index];
	}
	return password;
}

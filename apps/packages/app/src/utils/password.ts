export type GeneratePasswordOptions = {
	length: number;
	includeUppercase: boolean;
	includeLowercase: boolean;
	includeNumbers: boolean;
	includeSymbols: boolean;
};

export const generateRandomPassword = (
	options: GeneratePasswordOptions = {
		length: 16,
		includeUppercase: true,
		includeLowercase: true,
		includeNumbers: true,
		includeSymbols: true
	}
) => {
	const {
		length = 16,
		includeUppercase = true,
		includeLowercase = true,
		includeNumbers = true,
		includeSymbols = true
	} = options;

	let chars = '';

	if (includeLowercase) chars += 'abcdefghijklmnopqrstuvwxyz';
	if (includeUppercase) chars += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
	if (includeNumbers) chars += '0123456789';
	if (includeSymbols) chars += '!@#$%^&*()_+-=[]{}|;:,.<>?';

	let password = '';
	const array = new Uint8Array(length);
	crypto.getRandomValues(array);

	for (let i = 0; i < length; i++) {
		password += chars[array[i] % chars.length];
	}

	if (includeUppercase && !/[A-Z]/.test(password)) {
		password = replaceRandomChar(password, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ');
	}
	if (includeLowercase && !/[a-z]/.test(password)) {
		password = replaceRandomChar(password, 'abcdefghijklmnopqrstuvwxyz');
	}
	if (includeNumbers && !/[0-9]/.test(password)) {
		password = replaceRandomChar(password, '0123456789');
	}
	if (includeSymbols && !/[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]/.test(password)) {
		password = replaceRandomChar(password, '!@#$%^&*()_+-=[]{}|;:,.<>?');
	}

	return password;
};

const replaceRandomChar = (password: string, chars: string) => {
	const randomIndex = Math.floor(Math.random() * password.length);
	const randomChar = chars[Math.floor(Math.random() * chars.length)];
	return (
		password.substring(0, randomIndex) +
		randomChar +
		password.substring(randomIndex + 1)
	);
};

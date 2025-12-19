export class AddressValidator {
	static isValidAddress(address: string): boolean {
		return /^0x[a-fA-F0-9]{40}$/.test(address);
	}
}

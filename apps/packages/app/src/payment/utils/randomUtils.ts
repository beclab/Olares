export class RandomUtils {
	private static readonly NONCE_MIN = 10000000;
	private static readonly NONCE_MAX = 99999999;

	/**
	 * Generate a random nonce (8-digit number)
	 * @returns A random 8-digit number as a string
	 */
	static generateNonce(): string {
		const range = RandomUtils.NONCE_MAX - RandomUtils.NONCE_MIN + 1;
		return Math.floor(RandomUtils.NONCE_MIN + Math.random() * range).toString();
	}
}

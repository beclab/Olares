export interface NFTTokenData {
	animation_url: string;
	attributes: [{ trait_type: string; value: string }];
	background_color: string | null;
	description: string | null;
	external_url: string | null;
	image: string;
	name: string;
	timestamp: number;
	token_id: string;
	compiler?: string;
	date?: string;
	dna?: string;
	_id: string;
	contract_id: string;
}

export interface NFTItem {
	chain: string;
	tokenDatas: [NFTTokenData];
	token_meta: {
		address: string;
		name: string;
		source_type: string | null;
		symbol: string;
		type: string;
	};
	tokens: [
		{
			id: string;
			value: string;
			price: string;
		}
	];
}

export interface ProfileNFT {
	chain: string;
	token_name: string;
	token_address: string;
	token_id: string;
	nft_name: string;
	image: string;
	contract_id: string;
}

export interface AuthenticatedAddress {
	index: number;
	address: string;
}

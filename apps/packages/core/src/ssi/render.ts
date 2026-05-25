export interface ImageResource {
	uri: string;
	alt: string;
}

export interface ColorResource {
	color: string;
}

// EntityStyleDescriptor https://identity.foundation/wallet-rendering/#entity-styles
export interface EntityStyleDescriptor {
	thumbnail?: ImageResource;
	hero?: ImageResource;
	background?: ColorResource;
	text?: ColorResource;
}

// DataDisplay https://identity.foundation/wallet-rendering/#data-display
export interface DataDisplay {
	title?: DisplayMappingObject;
	subtitle?: DisplayMappingObject;
	description?: DisplayMappingObject;
	properties?: LabeledDisplayMappingObject[];
}

export interface DisplayMappingObject {
	path: string[];
	schema: DisplayMappingSchema;
	fallback: string;
	text: string;
}

export enum SchemaType {
	StringType = 'string',
	BooleanType = 'boolean',
	NumberType = 'number',
	IntegerType = 'integer'
}

export enum SchemaFormat {
	DateTimeFormat = 'date-time',
	TimeFormat = 'time',
	DateFormat = 'date',
	EmailFormat = 'email',
	IDNEmailFormat = 'idn-email',
	HostnameFormat = 'hostname',
	IDNHostnameFormat = 'idn-hostname',
	IPV4Format = 'ipv4',
	IPV6Format = 'ipv6',
	URIFormat = 'uri',
	URIReferenceFormat = 'uri-reference',
	IRIFormat = 'iri',
	IRIReferenceFormat = 'iri-reference'
}

export interface DisplayMappingSchema {
	type: SchemaType;
	format: SchemaFormat;
}

export interface LabeledDisplayMappingObject {
	type: SchemaType;
	format: SchemaFormat;
	label: string;
}

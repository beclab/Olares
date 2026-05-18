//type JSONSchema map[string]interface{}
//import { DataDisplay,EntityStyleDescriptor } from "./render";
import { JWT } from './keys';

export type JSONSchema = Record<string, any>;

// VCJSONSchema is the model representing the
// credential json schema specification https://w3c-ccg.github.io/vc-json-schemas/v2/index.html#credential_schema_definition
// export interface VCJSONSchema {
// 	type: string;
// 	version: string;
// 	id: string;
// 	name: string;
// 	author: string;
// 	authored: string;
// 	schema: JSONSchema;
// }

export interface CreateSchemaRequest {
	name: string;
	schema: JSONSchema;
	description?: string;
	issuer: string;
	verificationMethodId: string;
}

export interface CreateSchemaResponse {
	type: string;
	id: string;
	schema?: JSONSchema;
	credentialSchema?: JWT;
}

export interface GetSchemaResponse {
	type: string;
	id: string;
	schema?: JSONSchema;
	credentialSchema?: JWT;
	// schema: VCJSONSchema;
	// schemaJwt: JWT;
}

export interface GetSchemasResponse {
	schemas: GetSchemaResponse[];
}

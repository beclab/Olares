namespace go api.account

struct AccountReq {
  1: required string Name (api.body="name");
  2: required string Type (api.body="type");
  3: required string User (api.body="user");
}

struct AccountResp {
  1: i32 code;
  2: string message;
  3: AccountRespData data;
}

struct AccountRespData {
  1: string name;
  2: string type;
  3: AccountRespDataRaw rawData;
}

struct AccountRespDataRaw {
  1: string accessToken;
  2: string refreshToken;
  3: i64 expiresAt;
  4: i64 createAt;
  5: bool available;
  6: string endpoint (go.tag='json:"endpoint,omitempty"');
  7: string bucket (go.tag='json:"bucket,omitempty"');
  8: string userId (go.tag='json:"userId,omitempty"');
  9: string scope (go.tag='json:"scope,omitempty"');
  10: string idToken (go.tag='json:"idToken,omitempty"');
  11: string clientId (go.tag='json:"clientId,omitempty"');
  12: string cloudUrl (go.tag='json:"cloudUrl,omitempty"');
}

struct AccountListReq {
  1: required string User (api.body="user");
}

struct AccountListResp {
  1: i32 code;
  2: string message;
  3: list<AccountListData> data;
}

struct AccountListData {
  1: string name;
  2: string type;
  3: bool available;
  4: i64 createAt;
  5: i64 expiresAt;
}

service AccountService {
  AccountResp GetAccount(1: AccountReq request) (api.post="/api/account/retrieve");
  AccountListResp GetList(1: AccountListReq request) (api.post="/api/account/list");
}
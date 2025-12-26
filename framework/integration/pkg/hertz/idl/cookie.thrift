namespace go api.cookie

struct CookieReq {
  1: required string Domain (api.body="domain");
  2: required string User (api.body="user");
}

struct CookieResp {
  1: i32 code;
  2: string message;
  3: list<CookieRespData> data;
}

struct CookieRespData {
  1: string domain;
  2: string account;
  3: list<CookieRespDataRaw> records;
}

struct CookieRespDataRaw {
  1: string domain;
  2: string name;
  3: string value;
  4: double expires;
  5: string path;
  6: bool secure;
  7: bool httpOnly;
  8: string sameSite;
  9: string other;
}

service CookieService {
  CookieResp GetCookie(1: CookieReq request) (api.post="/api/cookie/retrieve");
}
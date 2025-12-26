local getenv = os.getenv
local env_auth_enabled = "AUTH_ENABLED"
local bfl_api_prefix = "http://127.0.0.1:8080"

local _M = {
    AUTH_ENABLED = getenv(env_auth_enabled) == "true" or false,

    BFL_AUTH_URL = bfl_api_prefix .. "/bfl/iam/v1alpha1/roles",

    BFL_BACKEND_USER_INFO = bfl_api_prefix .. "/bfl/backend/v1/user-info",

    COOKIE_TOKEN_KEY = "auth_token",

    BFL_AUTH_HEADER_KEY = "X-Authorization",

    CODE_SUCCESS = 0,

    CODE_TOKEN_VALIDATE_ERR = 100001,

    JSON_CONTENT_TYPE = "application/json; charset=utf-8",
}

-- uri white list
_M.NO_AUTH_URIs = { "/", "/login", "/bfl/iam/v1alpha1/login", "/bfl/backend/v1/user-info" }

-- add response headers
_M.ADD_HEADERS = {
    -- ["Content-Security-Policy"] = "upgrade-insecure-requests",
}

return _M


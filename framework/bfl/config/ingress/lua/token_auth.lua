local http = require("resty.http")
local cookie = require("resty.cookie")
local cjson = require("cjson.safe")
local constant = require("constant")

local ngx_req = ngx.req
local get_headers = ngx_req.get_headers
local set_req_header = ngx_req.set_header
local bfl = require("bfl_api")

local _M = {}

local function parse_token()
    -- parse token from http header
    local req_headers = get_headers()
    local header_token = req_headers["X-Authorization"]
    if header_token ~= nil and header_token ~= "" then
        return header_token, nil
    end

    local ck, err = cookie:new()
    if not ck then
        return nil, err
    end

    -- from http cookie
    local all_cookies = ck:get_all()
    if not all_cookies then
        return nil, "no http cookies"
    end

    ngx.log(ngx.INFO, "got all http cookies: ", cjson.encode(all_cookies))

    local cookie_token = all_cookies[constant.COOKIE_TOKEN_KEY]
    if cookie_token ~= nil or cookie_token ~= "" then
        return cookie_token, nil
    end

    return nil, "no auth token found"
end

local function send_request(token)
    local httpc = http.new()
    local res = {
        code = constant.CODE_TOKEN_VALIDATE_ERR,
        message = "",
    }

    -- setting timeouts
    httpc:set_timeouts(2000, 30000, 10000)

    local resp, err = httpc:request_uri(constant.BFL_AUTH_URL, {
        method = "GET",
        headers = {
            ["Content-Type"] = constant.JSON_CONTENT_TYPE,
            [constant.BFL_AUTH_HEADER_KEY] = token,
        }
    })

    if err then
        res.message = "validate token error: " .. err
        return res
    end

    local resp_body = resp.body
    ngx.log(ngx.INFO, "response status: ", resp.status, ", type: ", type(resp.body), ", body: " .. resp_body)

    local decoded_body = cjson.decode(resp_body)
    if not decoded_body then
        res.message = "unexpected token validate api response"
    end

    httpc:close()

    ngx.log(ngx.INFO, "decoded auth validation response: ", cjson.encode(decoded_body))
    res.code = decoded_body.code
    res.message = decoded_body.message

    return res
end

local function trust(uri)
    local trusted = false

    for _, v in pairs(constant.NO_AUTH_URIs) do
        if v == uri then
            trusted = true
            break
        end
    end

    return trusted
end

local function get_user()
    local res = {
        zone = "",
        name = ""
    }

    --local bflstore = ngx.shared.bflstore
    --local user_zone = bflstore:get("user_zone")
    --if not user_zone or user_zone == "" then
    --    ngx.log(ngx.ERR, "no user_zone data in ngx.shared.bflstore")
    --    return
    --end
    if not bfl_user_zone or bfl_user_zone == "" then
        ngx.log(ngx.ERR, "no bfl user_zone data")
        return
    end
    res.zone = bfl_user_zone

    if bfl_username and bfl_username ~= "" then
        res.name = bfl_username
    end

    ngx.log(ngx.INFO, "get bfl username: ", res.name, ", zone: ", res.zone)

    return res
end

local function filter_pass(host, user)
    -- for user desktop, no filter
    local trust_list = {
        string.format("profile-%s.%s", user.name, user.zone),
        string.format("wizard-%s.%s", user.name, user.zone),
        string.format("desktop-%s.%s", user.name, user.zone),
        string.format("desktop.%s", user.zone),
        string.format("auth-%s.%s", user.name, user.zone),
        string.format("auth.%s", user.zone),
        user.zone,
    }

    for _, item in ipairs(trust_list) do
        if item == host then
            return true
        end
    end

    return false
end

local function match_user_zone(host, user)
    local prefix = "^([a-zA-Z0-9-]*)(.?)"

    local reg = prefix .. user.zone .. "$"
    if host:match(reg) then
        return true
    end

    return false
end

local function verify_bfl_user(user)
    if user == nil then
        return false
    end

    if user.name == nil or user.name == "" or
        user.zone == nil or user.zone == "" then
        return false
    end
    return true
end

function _M.validate()
    if constant.AUTH_ENABLED == false then
        return
    end

    local user = bfl_user

    if user.name == nil or user.zone == nil then
        user = bfl.get_user_info()
        if verify_bfl_user(user) then
            ngx.log(ngx.INFO, "user not in cache, re fetched bfl user: ", cjson.encode(user))
            bfl_user = user
        end
    end

    ngx.log(ngx.INFO, "loaded bfl user: ", cjson.encode(bfl_user))

    if not verify_bfl_user(user) then
        ngx.say("no bfl user found in cache, try again later")
        ngx.status = ngx.HTTP_UNAUTHORIZED
        return ngx.exit(ngx.HTTP_UNAUTHORIZED)
    end

    -- public access policy
    if user.access_level ~= nil and user.access_level == 1 then
      ngx.log(ngx.INFO, "filter pass: public access policy")
      return
  end

    local host = ngx.var.http_host

    -- filter host pass, no auth
    if filter_pass(host, user) then
        ngx.log(ngx.INFO, "filter pass: host no auth")
        return
    end

    -- match user zone
    if not match_user_zone(host, user) then
        ngx.say("unknown server name, not matched the user's zone, attention!")
        ngx.status = ngx.HTTP_UNAUTHORIZED
        return ngx.exit(ngx.HTTP_UNAUTHORIZED)
    end

    -- pass options request
    local method = ngx_req.get_method()
    if method ~= nil and method == "OPTIONS" then
        ngx.log(ngx.INFO, "filter pass: OPTIONS request")
        return
    end

    -- validate token
    local res = {}
    local token, err = parse_token()

    if not token then
        res.code = constant.CODE_TOKEN_VALIDATE_ERR
        res.message = err
    else
        local d = send_request(token)

        res.code = d.code
        if not d.message then
            d.message = "token unauthorized, access denied"
        end
        res.message = d.message
    end

    if res.code ~= constant.CODE_SUCCESS then
        ngx.header["Content-Type"] = constant.JSON_CONTENT_TYPE
        ngx.status = ngx.HTTP_UNAUTHORIZED
        ngx.say(cjson.encode(res))

        ngx.log(ngx.ERR, res.message)
        return ngx.exit(ngx.HTTP_UNAUTHORIZED)
    end

    set_req_header(constant.BFL_AUTH_HEADER_KEY, token)
    ngx.log(ngx.INFO, "proxy_pass req headers: ", cjson.encode(get_headers()))
end

return _M

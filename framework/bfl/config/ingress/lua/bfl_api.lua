local http = require("resty.http")
local cjson = require("cjson.safe")
local constant = require("constant")
local tostring = tostring

local _M = {}

function _M.get_user_info()
    local httpc = http.new()
    httpc:set_timeouts(2000, 30000, 10000)

    local resp, err = httpc:request_uri(constant.BFL_BACKEND_USER_INFO, {
        method = "GET",
        headers = {
            ["Content-Type"] = constant.JSON_CONTENT_TYPE,
        }
    })

    if err ~= nil then
        ngx.log(ngx.ERR, "failed to request bfl backend user info: ", tostring(err))
        return
    end

    local res = cjson.decode(resp.body)
    if not res or (res.code ~= 0 and res.message) then
        ngx.log(ngx.ERR, string.format("response backend user info err, code: %s, message: %s",
                tostring(res.code), res.message))
        return
    end

    httpc:close()

    if res.data == nil then
        ngx.log(ngx.ERR, "unexpected response user info: not user data")
        return
    end

    return res.data
end

return _M
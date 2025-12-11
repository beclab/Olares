local cjson = require("cjson.safe")
local configuration_data = ngx.shared.tcp_udp_configuration_data
local tostring = tostring

local current_user_dict_sum = ""
local current_user_dict = {}

local _M = {}

-- set the nginx shared dict, it's shared memory for all nginx workers
function _M.set_users(users)
    local ok, err = configuration_data:set("users", users)
    if not ok then
        return "failed to updating users configuration, " .. tostring(err)
    end

    local ok err = configuration_data:set("users_timestamp", tostring(os.time()))
    if not ok then
        return "failed to updating users configuration md5 sum, " .. tostring(err)
    end

    return ""
end

-- get the users info from shared dict, every single worker must get the users info ownself
-- cause the worker process is isolated
local function get_users()
    -- check curent user dict sum, if empty or sum not equal, then get the users from shared dict
    user_sum = configuration_data:get("users_timestamp")
    if user_sum == "" then
        ngx.log(ngx.ERR, "get_users(): could not to get users data, users timestamp is empty")
        return
    end

    if current_user_dict_sum ~= "" and current_user_dict_sum == user_sum then
        return current_user_dict
    end

    local users_data, get_user_err = configuration_data:get("users")
    if not users_data then
        ngx.log(ngx.ERR, "get_users(): could not to get users data, " .. tostring(get_user_err))
        return
    end

    local users = cjson.decode(users_data)
    if not users then
        ngx.log(ngx.ERR, "get_users(): cjson.decode err, could not parse users data")
        return
    end

    current_user_dict = users
    current_user_dict_sum = user_sum

    return current_user_dict
end

function _M.list_users()
    return get_users()
end

function _M.get_user(name)
    if name == nil or name == "" then
        return
    end

    local users = get_users()
    if not users then
        return
    end

    for _, user in ipairs(users) do
        if name == user.name then
            return user
        end
    end
    return
end

return _M

local ngx = ngx
local cjson = require("cjson.safe")
local util = require("util")
local configuration = require("configuration")
local ipmatcher = require("resty.ipmatcher")
local ipairs = ipairs
local tbl_insert = table.insert

local _M = {}

function _M.run()
    local sock, err = ngx.req.socket(true)
    if not sock then
        ngx.log(ngx.ERR, "failed to get raw req socket: ", err)
        ngx.say("error: ", err)
        return
    end

    local reader = sock:receiveuntil("\r\n")
    local users, err_read = reader()
    if not users then
        ngx.log(ngx.ERR, "failed to read dynamic-configuration:", err_read)
        ngx.say("error: ", err_read)
        return
    end

    if users == nil or users == "" then
        return
    end

    local err_conf = configuration.set_users(users)
    if err_conf ~= "" then
        ngx.log(ngx.ERR, err_conf)
        ngx.say("error: ", err_conf)
        return
    end
end

local function match_user(server_name)
    -- load users
    local users = configuration.list_users()
    if not users then
        return
    end

    local ephemeral_users = {}
    local admin_users = {}

    for _, user in ipairs(users) do
        if user.is_ephemeral ~= nil and user.is_ephemeral == "yes" then
            tbl_insert(ephemeral_users, user)
        else
            tbl_insert(admin_users, user)
        end
    end

    ngx.log(ngx.INFO, "ephemeral users: ", cjson.encode(ephemeral_users))
    ngx.log(ngx.INFO, "admin users: ", cjson.encode(admin_users))

    local reg = ""

    -- ephemeral users
    for _, user in ipairs(ephemeral_users) do
        reg = string.format("^([A-Za-z0-9]+)-%s.%s$", user.name, user.zone)
        if server_name:match(reg) then
            ngx.log(ngx.INFO, "server_name: ", server_name, ", matched user: ", user.name)
            return user
        end

        -- local zone
        reg = string.format("^([A-Za-z0-9]+)-%s.olares.local$", user.name)
        if server_name:match(reg) then
            ngx.log(ngx.INFO, "server_name: ", server_name, ", matched user: ", user.name)
            return user
        end
    end

    -- admin users
    local prefix = "^([a-zA-Z0-9-]*)(.?)"
    for _, user in ipairs(admin_users) do
        for _, domain in ipairs(user.ngx_server_name_domains) do
            if server_name == domain then
                return user
            end

            domain_reg = string.gsub(domain, "%-", "%%-")
            reg = prefix .. domain_reg .. "$"
            if server_name:match(reg) then
                ngx.log(ngx.INFO, "server_name: ", server_name, ", matched user: ", user.name)
                return user
            end
        end
    end

    return nil
end

local function set_variables(user)
    ngx.var.bfl_username = user.name
    ngx.var.bfl_ingress_host = user.bfl_ingress_svc_host
    ngx.var.bfl_ingress_port = user.bfl_ingress_svc_port
end

local function deny_filter(user, server_name)
    -- always allow local ip
    local remote_addr = ngx.var.remote_addr
    if remote_addr == user.local_domain_ip then
        return true
    end

    -- allow vpn cidr
    local vpn_cidr = ipmatcher.new({"100.64.0.0/24"})
    if vpn_cidr:match(remote_addr) then
        return true
    end

    local deny = user.deny_all
    if deny == nil or deny == 0 then
        return true
    end

    local allowed_domains = user.allowed_domains
    if not allowed_domains then
        ngx.log(ngx.ERR, "deny all access domain")
        return false
    end

    for _, domain in ipairs(allowed_domains) do
        if domain == server_name then
            return true
        end
    end

    return false
end

local function access_filter(user)
    local remote_addr = ngx.var.remote_addr
    if not remote_addr or remote_addr == "" then
        ngx.log(ngx.ERR, "can not get remote_addr")
        return false
    end

    local addrs = {
        remote_addr = ngx.var.remote_addr,
        proxy_protocol_addr = ngx.var.proxy_protocol_addr,
        realip_remote_addr = ngx.var.realip_remote_addr
    }

    if addrs.proxy_protocol_addr and addrs.proxy_protocol_addr ~= "" then
        remote_addr = addrs.proxy_protocol_addr
    end

    ngx.log(ngx.INFO, string.format("client remote_addr: %s, and addrs: %s",
        remote_addr, cjson.encode(addrs)))

    -- network access level policy
    local access_level = user.access_level
    if not access_level or type(access_level) ~= "number" or access_level == 0 then
        ngx.log(ngx.ERR, "user ", user.name, ", no network access_level")
        return false
    end

    if access_level == 1 or access_level == 2 then
        return true
    end

    -- filter with user's allow_cidrs
    if user.allow_cidrs == nil then
        ngx.log(ngx.ERR, "user ", user.name, " has no allow_cidrs configuration")
        return false
    end

    ngx.log(ngx.DEBUG, "user ", user.name, ", allowed ip cidrs: " .. cjson.encode(user.allow_cidrs))

    local ip = ipmatcher.new(user.allow_cidrs)
    if not ip:match(remote_addr) then
        ngx.log(ngx.ERR, string.format(
            "client: %s, access forbidden!", remote_addr))
        return false
    end
    return true
end

function _M.preread()
    local server_name = ngx.var.ssl_preread_server_name

    if server_name == nil or server_name == "" then
        ngx.log(ngx.ERR, "no ssl server_name")
        return
    end

    ngx.log(ngx.INFO, "preread ssl server_name: " .. server_name)

    local curr_user = match_user(server_name)
    if not curr_user then
        ngx.log(ngx.ERR, "server name " .. server_name .. ", could not match any users")
        return ngx.exit(400)
    end

    ngx.log(ngx.INFO, "server_name: " .. server_name .. ", current user: " .. cjson.encode(curr_user))

    -- set user variables
    set_variables(curr_user)

    -- access filter
    -- if not access_filter(curr_user) then
    --     return ngx.exit(403)
    -- end

    -- deny filter
    if not deny_filter(curr_user, server_name) then
        return ngx.exit(403)
    end
end

setmetatable(_M, {
    __index = {
        match_user = match_user,
    }
})

return _M

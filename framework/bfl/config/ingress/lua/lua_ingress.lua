local util = require("util")
local cjson = require("cjson.safe")
local constant = require("constant")
local bfl = require("bfl_api")

local _M = {}

local function sync_bfl_user()
    if bfl_user.name ~= nil and bfl_user.zone ~= nil then
        ngx.log(ngx.DEBUG, "sync: bfl user is loaded, ignore")
        return
    end

    local user = bfl.get_user_info()
    if user ~= nil and user.name ~= nil and user.zone ~= nil then
        ngx.log(ngx.INFO, "sync: successfully fetch bfl user: ", cjson.encode(user))
        bfl_user = user
    end
end

function _M.init_worker()
    ngx.log(ngx.INFO, "init worker, new every timer(7s) to load bfl user info")

    local ok, err = ngx.timer.every(7, sync_bfl_user)
    if not ok then
        ngx.log(ngx.ERR, "failed to create sync_bfl_user timer: ", err)
    end
end

function _M.force_to_https()
    local uri = ngx.var.request_uri
    local scheme, server_name = ngx.var.scheme, ngx.var.server_name

    if scheme ~= "https" then
       ngx.status = ngx.HTTP_MOVED_PERMANENTLY
       ngx.redirect("https://" .. server_name .. uri, ngx.HTTP_MOVED_PERMANENTLY)
    end
 end

function _M.set_req_user()
    local user = bfl.get_user_info()
    if user == nil then
        ngx.log(ngx.ERR, "failed to get bfl user")
        return
    end

    if user.name ~= "" then
        ngx.req.set_header("X-BFL-USER", user.name)
        return
    end
end

function _M.add_response_headers()
    if not constant.ADD_HEADERS then
        return
    end

    for k, v in pairs(constant.ADD_HEADERS) do
        ngx.header[k] = v
    end
end

function _M.overwrite_response_to_https()
    local loc = ngx.header.Location
    if not loc then
       return
    end

    local t = util.string_split(loc, "://")
    if not t or #t ~= 2 then
        ngx.log(ngx.ERR, "split location " .. loc .. ", got unexpected value")
        return
    end

    local scheme, url = t[1], t[2]
    ngx.log(ngx.INFO, "parsed url, scheme: ", scheme, ", url: ", url)
    if scheme ~= nil and url ~= nil then
        local domain = util.string_split(url, "/")
        local domainSuffix = "(.*).olares.local$"
        if scheme == "http" and url ~= "" and not domain[1]:match(domainSuffix) then
            ngx.header["Location"] = "https://" .. url
        end
    end
end

return _M

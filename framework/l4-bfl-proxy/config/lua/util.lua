local tbl_insert = table.insert

local _M = {
    _VERSION = "0.1",
}

function _M.string_split(str, pat)
    local t = {}  -- NOTE: use {n = 0} in Lua-5.0
    local fpat = "(.-)" .. pat
    local last_end = 1
    local s, e, cap = str:find(fpat, 1)
    while s do
       if s ~= 1 or cap ~= "" then
          tbl_insert(t, cap)
       end
       last_end = e+1
       s, e, cap = str:find(fpat, last_end)
    end
    if last_end <= #str then
       cap = str:sub(last_end)
       tbl_insert(t, cap)
    end

    return t
 end

function _M.table_to_array(obj)
   local t = {}

   if obj == nil then
      return t
   end

   for k, v in pairs(obj) do
         t[k] = v
   end
   return t
end

function _M.table_exists(t, v)
    for _, item in ipairs(t) do
        if item == v then
            return true
        end
    end

    return false
end

return _M

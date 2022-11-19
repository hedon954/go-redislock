local val = redis.call("get", KEYS[1])
if val == false then
    -- key not exists
    return redis.call("set", KEYS[1], ARGV[1], 'EX', ARGV[2])
elseif val == ARGV[1] then
    -- lock successfully last time, reset expiration
    redis.call("expire", KEYS[1], ARGV[2])
    return "OK"
else
    -- lock by others
    return ""
end
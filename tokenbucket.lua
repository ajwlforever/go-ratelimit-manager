--[[
  1. key - 令牌桶的 key
  2. intervalPerTokens - 生成令牌的间隔(ms)
  3. curTime - 当前时间
  4. initTokens - 令牌桶初始化的令牌数
  5. bucketMaxTokens - 令牌桶的上限
  6. resetBucketInterval - 重置桶内令牌的时间间隔
  7. currentTokens - 当前桶内令牌数
  8. bucket - 当前 key 的令牌桶对象
]] --

local key = KEYS[1]
local intervalPerTokens = tonumber(ARGV[1])
local curTime = tonumber(ARGV[2])
local initTokens = tonumber(ARGV[3])
local bucketMaxTokens = tonumber(ARGV[4])
local resetBucketInterval = tonumber(ARGV[5])

local bucket = redis.call('hgetall', key)
local currentTokens

-- 若当前桶未初始化,先初始化令牌桶
if table.maxn(bucket) == 0 then
    -- 初始桶内令牌
    currentTokens = initTokens
    -- 设置桶最近的填充时间是当前
    redis.call('hset', key, 'lastRefillTime', curTime)
    -- 初始化令牌桶的过期时间, 设置为间隔的 1.5 倍
    redis.call('pexpire', key, resetBucketInterval * 1.5)

-- 若桶已初始化,开始计算桶内令牌
-- 为什么等于 4 ? 因为有两对 field, 加起来长度是 4
-- { "lastRefillTime(上一次更新时间)","curTime(更新时间值)","tokensRemaining(当前保留的令牌)","令牌数" }
elseif table.maxn(bucket) == 4 then

    -- 上次填充时间
    local lastRefillTime = tonumber(bucket[2])
    -- 剩余的令牌数
    local tokensRemaining = tonumber(bucket[4])

    -- 当前时间大于上次填充时间
    if curTime > lastRefillTime then

        -- 拿到当前时间与上次填充时间的时间间隔
        -- 举例理解: curTime = 2620 , lastRefillTime = 2000, intervalSinceLast = 620
        local intervalSinceLast = curTime - lastRefillTime

        -- 如果当前时间间隔 大于 令牌的生成间隔
        -- 举例理解: intervalSinceLast = 620, resetBucketInterval = 1000
        if intervalSinceLast > resetBucketInterval then

            -- 将当前令牌填充满
            currentTokens = initTokens

            -- 更新重新填充时间
            redis.call('hset', key, 'lastRefillTime', curTime)
            
        -- 如果当前时间间隔 小于 令牌的生成间隔
        else

            -- 可授予的令牌 = 向下取整数( 上次填充时间与当前时间的时间间隔 / 两个令牌许可之间的时间间隔 )
            -- 举例理解 : intervalPerTokens = 200 ms , 令牌间隔时间为 200ms
            --           intervalSinceLast = 620 ms , 当前距离上一个填充时间差为 620ms
            --           grantedTokens = 620/200 = 3.1 = 3
            local grantedTokens = math.floor(intervalSinceLast / intervalPerTokens)

            -- 可授予的令牌 > 0 时
            -- 举例理解 : grantedTokens = 620/200 = 3.1 = 3
            if grantedTokens > 0 then

                -- 生成的令牌 = 上次填充时间与当前时间的时间间隔 % 两个令牌许可之间的时间间隔
                -- 举例理解 : padMillis = 620%200 = 20
                --           curTime = 2620
                --           curTime - padMillis = 2600
                local padMillis = math.fmod(intervalSinceLast, intervalPerTokens)

                -- 将当前令牌桶更新到上一次生成时间
                redis.call('hset', key, 'lastRefillTime', curTime - padMillis)
            end

            -- 更新当前令牌桶中的令牌数
            -- Math.min(根据时间生成的令牌数 + 剩下的令牌数, 桶的限制) => 超出桶最大令牌的就丢弃
            currentTokens = math.min(grantedTokens + tokensRemaining, bucketMaxTokens)
        end
    else
        -- 如果当前时间小于或等于上次更新的时间, 说明刚刚初始化, 当前令牌数量等于桶内令牌数
        -- 不需要重新填充
        currentTokens = tokensRemaining
    end
end

-- 如果当前桶内令牌小于 0,抛出异常
assert(currentTokens >= 0)

-- 如果当前令牌 == 0 ,更新桶内令牌, 返回 0
if currentTokens == 0 then
    redis.call('hset', key, 'tokensRemaining', currentTokens)
    return 0
else
    -- 如果当前令牌 大于 0, 更新当前桶内的令牌 -1 , 再返回当前桶内令牌数
    redis.call('hset', key, 'tokensRemaining', currentTokens - 1)
    return currentTokens
end
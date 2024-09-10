local crypto = require("crypto")
assert(crypto.sha256("lua") == "dc436b329f4da6c88b0b6ce79d829ac4fc33746b454604e2bcbad25a8e2985fe")

-- remove "not" from the following line to test http module
if not true then
	local http = require("http")
	local res = http.get("https://ip.me", {
		headers = { ["user-agent"] = "curl/0.0.0" },
	})
	local ip = res.body:gsub("%s+", "")
	print(ip)
end

local json = require("json")
assert(json.encode({ a = 1 }) == [[{"a":1}]])
assert(json.decode([[{"a":1}]]).a == 1)

local logger = require("logger")
logger.info("message from script", { foo = "bar", baz = 1 })

local re = require("re")
local start, finish = re.find("hello world", "o w", 1, true)
assert(start == 5 and finish == 7)
assert(re.gsub("hello world", "world", "WORLD") == "hello WORLD")
assert(re.match("hello world", "\\w+") == "hello")

local url = require("url")
local parsed = url.parse("https://httpbingo.org/get")
parsed.query = url.build_query_string({ a = "b" })
assert(url.build(parsed) == "https://httpbingo.org/get?a=b")

return true

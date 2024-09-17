local json = require("json")

local m = require("lmb")
m.state.status_code = 418
m.state.headers = {
	["content-type"] = "application/json",
	["x-appearence"] = { "material=pottery", "culture=Chinese" },
	["x-i-am"] = "a teapot",
}

return json.encode({ bool = true, num = 1, str = "str" })

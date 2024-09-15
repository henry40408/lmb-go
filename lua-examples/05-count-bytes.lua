local io = require("io")
local subtotal = {}

local function string_to_table(input)
	if type(input) == "table" then
		return input
	end

	local bytes = {}
	for i = 1, #input do
		table.insert(bytes, string.byte(input, i))
	end

	return bytes
end

while true do
	local buf = io.read(1000)
	if not buf then
		break
	end
	local bytes = string_to_table(buf)
	if not bytes then
		break
	end
	for _, b in pairs(bytes) do
		subtotal[b] = (subtotal[b] or 0) + 1
	end
end

local total = 0
for b, count in pairs(subtotal) do
	print(b, ":", count)
	total = total + count
end

print("total:", total)
return true

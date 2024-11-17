# Lua Guide

## Language Variant and Version

Lmb currently uses GopherLua.

```lua
assert('Lua 5.1' == _G._VERSION, 'expect ' .. _G._VERSION) -- Lua version
```

GopherLua is a Lua 5.1 (+goto statement in Lua 5.2) VM and compiler written in Go.

## Hello, world

First things first: Hello, World!

Save the following file as "hello.lua".

```lua
print('hello, world!')
```

Run the script:

```sh
$ lmb eval --file hello.lua
hello, world!
```

## I/O library

For security, the original `io` library is removed. However, because it's common to print and read something in daily use, Lmb implements the following functions/methods:

```lua
local io = require('io')

-- https://www.lua.org/manual/5.1/manual.html#pdf-print
print('hello, world!')

-- https://www.lua.org/manual/5.1/manual.html#pdf-io.stdout
io.write('standard output')

-- https://www.lua.org/manual/5.1/manual.html#pdf-io.read
assert(io.read('*L') == nil, 'expect nil for *L')
assert(io.read('*a') == nil, 'expect nil for *a')
assert(io.read('*l') == nil, 'expect nil for *l')
assert(io.read('*n') == nil, 'expect nil for *n')

local _, err = pcall(function() return io.read('*x') end)
assert(tostring(err):find('unsupported string format'))

-- https://www.lua.org/manual/5.1/manual.html#pdf-io.stderr
io.stderr:write('standard error')

-- output: standard output
```

```lua
-- input: 世界
return require('io').read(3)
-- output: 世
```

```lua
-- input: foobar
return require('io').read('*a')
-- output: foobar
```

```lua
-- input: 1949
return require('io').read('*n') ^ 2
-- output: 3798601
```

```lua
-- input: foo\nbar
return require('io').read('*l')
-- output: foo
```

```lua
-- input: foo\nbar
return require('io').read('*L')
-- output: foo\n
```

## Store

Lmb supports a key-value store backed by SQLite. The data can be read, written, and updated using the following APIs:

```lua
local m = require('@lmb')

-- read value from store
assert(not m.store['a'])

-- write value to store
m.store['a'] = 1
assert(m.store['a'] == 1)

-- update values in a transaction
function transfer(amount)
  m.store:update(function(store)
    if store['alice'] < amount then
      error('insufficient fund')
    end
    store['alice'] = store['alice'] - amount
    store['bob'] = store['bob'] + amount
  end)
end

-- failure scenario
m.store['alice'] = 50
m.store['bob'] = 50
local _, err = pcall(function()
  transfer(100)
end)
assert(tostring(err):find('insufficient fund')) -- error occurred
assert(m.store['alice'] == 50) -- as is
assert(m.store['bob'] == 50) -- as is

-- success scenario
m.store['alice'] = 100
m.store['bob'] = 0
local _, err = pcall(function()
  transfer(100)
end)
assert(not err) -- no error occurred
assert(m.store['alice'] == 0) -- withdrawn
assert(m.store['bob'] == 100) -- deposited
```

## HTTP `http`

Lmb is able to send HTTP requests. The following example sends a GET request to https://httpbin.org/headers with the header `I-Am: A teapot`:

```lua
local http = require('http')
local json = require('json')

local res, err = http.get('https://httpbingo.org/headers', {
  headers = {
    ['I-Am'] = 'A teapot',
  },
})
assert(not err, tostring(err))
assert('application/json' == res.headers['Content-Type'])

local parsed = json.decode(res.body)
assert('A teapot' == parsed['headers']['I-Am'], parsed)
```

## JSON `json`

JSON is a common format used to send HTTP requests. Lmb supports both encoding and decoding JSON data:

```lua
local json = require('json')
assert('{"bool":true,"num":1.23,"str":"string"}' == json.encode({ bool = true, num = 1.23, str = 'string' }))
assert('[true,1.23,"string"]' == json.encode({ true, 1.23, 'string' }))

local decoded = json.decode('{"bool":true,"num":1.23,"str":"string"}')
assert(true == decoded.bool)
assert(1.23 == decoded.num)
assert('string' == decoded.str)

local decoded = json.decode('[true,1.23,"string"]')
assert(true == decoded[1])
assert(1.23 == decoded[2])
assert('string' == decoded[3])
```

## Cryptography `crypto`

When receiving webhook events from another service, e.g. [GitHub](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries), it's secure to validate them before processing.

```lua
local crypto = require('crypto')
assert(crypto.base64_encode('') == '')
assert(crypto.crc32('') == '00000000')
assert(crypto.md5('') == 'd41d8cd98f00b204e9800998ecf8427e')
assert(crypto.sha1('') == 'da39a3ee5e6b4b0d3255bfef95601890afd80709')
assert(crypto.sha256('') == 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855')
assert(crypto.sha512('') == 'cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e')
assert(crypto.hmac('sha256', '', 'secret') == 'f9e66e179b6747ae54108f82f8ade8b3c25d76fd30afde6c395822c530196169')
```

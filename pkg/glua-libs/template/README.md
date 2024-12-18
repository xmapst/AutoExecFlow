# template

## Usage

```lua
local template = require("template")

local mustache, err = template.choose("mustache")

local values = {name="world"}
print( mustache:render("Hello {{name}}!", values) ) -- mustache:render_file()
-- Output:"Hello world!"

local values = {data = {"one", "two"}}
print( mustache:render("{{#data}} {{.}} {{/data}}", values) )
-- Output:" one two "
```

## Supported engines

* [mustache](https://mustache.github.io/) [cbroglie/mustache](https://github.com/cbroglie/mustache)


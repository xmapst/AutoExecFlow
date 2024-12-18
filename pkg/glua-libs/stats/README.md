# stats

## Usage

```lua
local stats = require("stats")

local result, _ = stats.median({0,0,10})
print(result)
-- Output: 0

local result, _ = stats.percentile({0,0,10}, 100)
print(result)
-- Output: 10
```


# yaegi 

exec yaegi type script

## Usage

```text
# params is dict
import (
    "context"
    "fmt"
    
    "github.com/tidwall/gjson"
)

func EvalCall(ctx context.Context, params gjson.Result) {
	fmt.Println(params)
	fmt.Println("Hello, World!")
}
```
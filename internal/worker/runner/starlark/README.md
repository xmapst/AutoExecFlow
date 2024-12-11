# starlark 

exec starlark type script

## Usage

```text
# params is dict
def EvalCall(params):
    print(params)
    coins = {
      'dime': 10,
      'nickel': 5,
      'penny': 1,
      'quarter': 25,
    }
    print('By name:\t' + ', '.join(sorted(coins.keys())))
    print('By value:\t' + ', '.join(sorted(coins.keys(), key=coins.get)))
```
// glua with preload libs
// original: https://raw.githubusercontent.com/yuin/gopher-lua/master/cmd/glua/glua.go
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/chzyer/readline"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"

	libs "github.com/xmapst/AutoExecFlow/pkg/glua-libs"
)

func main() {
	os.Exit(mainAux())
}

func mainAux() int {
	var optE, optL, optP string
	var optI, optV, optDT, optDC bool
	var optM, optRS int
	flag.StringVar(&optE, "e", "", "")
	flag.StringVar(&optL, "l", "", "")
	flag.StringVar(&optP, "p", "", "")
	flag.IntVar(&optM, "mx", 0, "")
	flag.BoolVar(&optI, "i", false, "")
	flag.BoolVar(&optV, "v", false, "")
	flag.BoolVar(&optDT, "dt", false, "")
	flag.BoolVar(&optDC, "dc", false, "")
	flag.IntVar(&optRS, "r", lua.RegistrySize, "")
	flag.Usage = func() {
		fmt.Printf(`Usage: glua-libs [options] [script [args]].
Available options are:
  -e stat  execute string 'stat'
  -l name  require library 'name'
  -mx MB   memory limit(default: unlimited)
  -dt      dump AST trees
  -dc      dump VM codes
  -r       registry size, default: %d
  -i       enter interactive mode after executing 'script'
  -p file  write cpu profiles to the file
  -v       show version information %s`, lua.RegistrySize, "\n")
	}
	flag.Parse()
	if len(optP) != 0 {
		f, err := os.Create(optP)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if len(optE) == 0 && !optI && !optV && flag.NArg() == 0 {
		optI = true
	}

	status := 0

	lua.RegistrySize = optRS
	L := lua.NewState()
	libs.Preload(L)
	defer L.Close()
	if optM > 0 {
		L.SetMx(optM)
	}

	if optV || optI {
		fmt.Println(lua.PackageCopyRight)
	}

	if len(optL) > 0 {
		if err := L.DoFile(optL); err != nil {
			fmt.Println(err.Error())
		}
	}

	if nargs := flag.NArg(); nargs > 0 {
		script := flag.Arg(0)
		argtb := L.NewTable()
		for i := 1; i < nargs; i++ {
			L.RawSet(argtb, lua.LNumber(i), lua.LString(flag.Arg(i)))
		}
		L.SetGlobal("arg", argtb)
		if optDT || optDC {
			file, err := os.Open(script)
			if err != nil {
				fmt.Println(err.Error())
				return 1
			}
			chunk, err2 := parse.Parse(file, script)
			if err2 != nil {
				fmt.Println(err2.Error())
				return 1
			}
			if optDT {
				fmt.Println(parse.Dump(chunk))
			}
			if optDC {
				proto, err3 := lua.Compile(chunk, script)
				if err3 != nil {
					fmt.Println(err3.Error())
					return 1
				}
				fmt.Println(proto.String())
			}
		}
		if err := L.DoFile(script); err != nil {
			fmt.Println(err.Error())
			status = 1
		}
	}

	if len(optE) > 0 {
		if err := L.DoString(optE); err != nil {
			fmt.Println(err.Error())
			status = 1
		}
	}

	if optI {
		doREPL(L)
	}
	return status
}

// do read/eval/print/loop
func doREPL(L *lua.LState) {
	rl, err := readline.NewEx(&readline.Config{Prompt: "> "})
	if err != nil {
		panic(err)
	}
	defer rl.Close()
	for {
		if str, err := loadline(rl, L); err == nil {
			if err := L.DoString(str); err != nil {
				fmt.Println(err)
			}
		} else { // error on loadline
			fmt.Println(err)
			return
		}
	}
}

func incomplete(err error) bool {
	var lerr *lua.ApiError
	if errors.As(err, &lerr) {
		var perr *parse.Error
		if errors.As(lerr.Cause, &perr) {
			return perr.Pos.Line == parse.EOF
		}
	}
	return false
}

func loadline(rl *readline.Instance, L *lua.LState) (string, error) {
	rl.SetPrompt("> ")
	if line, err := rl.Readline(); err == nil {
		if _, err := L.LoadString("return " + line); err == nil { // try add return <...> then compile
			return line, nil
		} else {
			return multiline(line, rl, L)
		}
	} else {
		return "", err
	}
}

func multiline(ml string, rl *readline.Instance, L *lua.LState) (string, error) {
	for {
		if _, err := L.LoadString(ml); err == nil { // try compile
			return ml, nil
		} else if !incomplete(err) { // syntax error , but not EOF
			return ml, nil
		} else {
			rl.SetPrompt(">> ")
			if line, err := rl.Readline(); err == nil {
				ml = ml + "\n" + line
			} else {
				return "", err
			}
		}
	}
}

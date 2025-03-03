// Code generated by 'yaegi extract github.com/panjf2000/ants/v2'. DO NOT EDIT.

package libs

import (
	"github.com/panjf2000/ants/v2"
	"go/constant"
	"go/token"
	"reflect"
)

func init() {
	Symbols["github.com/panjf2000/ants/v2/ants"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"CLOSED":                          reflect.ValueOf(constant.MakeFromLiteral("1", token.INT, 0)),
		"Cap":                             reflect.ValueOf(ants.Cap),
		"DefaultAntsPoolSize":             reflect.ValueOf(constant.MakeFromLiteral("2147483647", token.INT, 0)),
		"DefaultCleanIntervalTime":        reflect.ValueOf(ants.DefaultCleanIntervalTime),
		"ErrInvalidLoadBalancingStrategy": reflect.ValueOf(&ants.ErrInvalidLoadBalancingStrategy).Elem(),
		"ErrInvalidMultiPoolSize":         reflect.ValueOf(&ants.ErrInvalidMultiPoolSize).Elem(),
		"ErrInvalidPoolExpiry":            reflect.ValueOf(&ants.ErrInvalidPoolExpiry).Elem(),
		"ErrInvalidPoolIndex":             reflect.ValueOf(&ants.ErrInvalidPoolIndex).Elem(),
		"ErrInvalidPreAllocSize":          reflect.ValueOf(&ants.ErrInvalidPreAllocSize).Elem(),
		"ErrLackPoolFunc":                 reflect.ValueOf(&ants.ErrLackPoolFunc).Elem(),
		"ErrPoolClosed":                   reflect.ValueOf(&ants.ErrPoolClosed).Elem(),
		"ErrPoolOverload":                 reflect.ValueOf(&ants.ErrPoolOverload).Elem(),
		"ErrTimeout":                      reflect.ValueOf(&ants.ErrTimeout).Elem(),
		"Free":                            reflect.ValueOf(ants.Free),
		"LeastTasks":                      reflect.ValueOf(ants.LeastTasks),
		"NewMultiPool":                    reflect.ValueOf(ants.NewMultiPool),
		"NewMultiPoolWithFunc":            reflect.ValueOf(ants.NewMultiPoolWithFunc),
		"NewPool":                         reflect.ValueOf(ants.NewPool),
		"NewPoolWithFunc":                 reflect.ValueOf(ants.NewPoolWithFunc),
		"OPENED":                          reflect.ValueOf(constant.MakeFromLiteral("0", token.INT, 0)),
		"Reboot":                          reflect.ValueOf(ants.Reboot),
		"Release":                         reflect.ValueOf(ants.Release),
		"ReleaseTimeout":                  reflect.ValueOf(ants.ReleaseTimeout),
		"RoundRobin":                      reflect.ValueOf(ants.RoundRobin),
		"Running":                         reflect.ValueOf(ants.Running),
		"Submit":                          reflect.ValueOf(ants.Submit),
		"WithDisablePurge":                reflect.ValueOf(ants.WithDisablePurge),
		"WithExpiryDuration":              reflect.ValueOf(ants.WithExpiryDuration),
		"WithLogger":                      reflect.ValueOf(ants.WithLogger),
		"WithMaxBlockingTasks":            reflect.ValueOf(ants.WithMaxBlockingTasks),
		"WithNonblocking":                 reflect.ValueOf(ants.WithNonblocking),
		"WithOptions":                     reflect.ValueOf(ants.WithOptions),
		"WithPanicHandler":                reflect.ValueOf(ants.WithPanicHandler),
		"WithPreAlloc":                    reflect.ValueOf(ants.WithPreAlloc),

		// type definitions
		"LoadBalancingStrategy": reflect.ValueOf((*ants.LoadBalancingStrategy)(nil)),
		"Logger":                reflect.ValueOf((*ants.Logger)(nil)),
		"MultiPool":             reflect.ValueOf((*ants.MultiPool)(nil)),
		"MultiPoolWithFunc":     reflect.ValueOf((*ants.MultiPoolWithFunc)(nil)),
		"Option":                reflect.ValueOf((*ants.Option)(nil)),
		"Options":               reflect.ValueOf((*ants.Options)(nil)),
		"Pool":                  reflect.ValueOf((*ants.Pool)(nil)),
		"PoolWithFunc":          reflect.ValueOf((*ants.PoolWithFunc)(nil)),

		// interface wrapper definitions
		"_Logger": reflect.ValueOf((*_github_com_panjf2000_ants_v2_Logger)(nil)),
	}
}

// _github_com_panjf2000_ants_v2_Logger is an interface wrapper for Logger type
type _github_com_panjf2000_ants_v2_Logger struct {
	IValue  interface{}
	WPrintf func(format string, args ...any)
}

func (W _github_com_panjf2000_ants_v2_Logger) Printf(format string, args ...any) {
	W.WPrintf(format, args...)
}

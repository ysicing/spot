package flags

import (
	"github.com/ergoapi/util/zos"
	flag "github.com/spf13/pflag"
)

func GetDefaultConfig() string {
	return zos.GetHomeDir() + "/.spot.yaml"
}

// GlobalFlags is the flags that contains the global flags
type GlobalFlags struct {
	Debug      bool
	Silent     bool
	ConfigPath string
	Flags      *flag.FlagSet
}

// SetGlobalFlags applies the global flags
func SetGlobalFlags(flags *flag.FlagSet) *GlobalFlags {
	globalFlags := &GlobalFlags{
		Flags: flags,
	}
	flags.BoolVar(&globalFlags.Debug, "debug", false, "Prints the stack trace if an error occurs")
	flags.StringVar(&globalFlags.ConfigPath, "config", GetDefaultConfig(), "The ergo config file to use")
	return globalFlags
}

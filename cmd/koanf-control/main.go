package main

/*
- Provide a base command with standard flags
- Build child commands over top with additional flags
- Pull like flags from either CLI or env
- Standard prefixes to allow a shared libary of standard ochami flags? or at least some means of merging in an external flag set

Currently CI wants to use th original flags and vars

bruv 80% of https://pkg.go.dev/github.com/knadh/koanf#section-readme is whatever you should force people to name their tomls toml
*/

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

// rootCommand.Flags().StringVar(&pcs.stateManagerServer, "sms-server", defaultSMSServer, "SMS Server")
// rootCommand.Flags().BoolVar(&pcs.runControl, "run-control", pcs.runControl, "run control loop; false runs API only") //this was a flag useful for dev work
// rootCommand.Flags().BoolVar(&pcs.hsmLockEnabled, "hsmlock-enabled", true, "Use HSM Locking")                         // This was a flag useful for dev work
// rootCommand.Flags().BoolVar(&pcs.vaultEnabled, "vault-enabled", true, "Should vault be used for credentials?")
// rootCommand.Flags().StringVar(&pcs.vaultKeypath, "vault-keypath", "secret/hms-creds",
// 	"Keypath for Vault credentials.")
// rootCommand.Flags().IntVar(&pcs.credCacheDuration, "cred-cache-duration", 600,
// 	"Duration in seconds to cache vault credentials.")
//
// rootCommand.Flags().IntVar(&pcs.maxNumCompleted, "max-num-completed", defaultMaxNumCompleted, "Maximum number of completed records to keep.")
// rootCommand.Flags().IntVar(&pcs.expireTimeMins, "expire-time-mins", defaultExpireTimeMins, "The time, in mins, to keep completed records.")
//
// // ETCD flags
// rootCommand.Flags().BoolVar(&etcd.disableSizeChecks, "etcd-disable-size-checks", false, "Disables checking object size before storing and doing message truncation and paging.")
// rootCommand.Flags().IntVar(&etcd.pageSize, "etcd-page-size", storage.DefaultEtcdPageSize, "The maximum number of records to put in each etcd entry.")
// rootCommand.Flags().IntVar(&etcd.maxMessageLength, "etcd-max-transition-message-length", storage.DefaultMaxMessageLen, "The maximum length of messages per task in a transition.")
// rootCommand.Flags().IntVar(&etcd.maxObjectSize, "etcd-max-object-size", storage.DefaultMaxEtcdObjectSize, "The maximum data size in bytes for objects in etcd.")
//
// // JWKS URL flag
// rootCommand.Flags().StringVar(&jwksURL, "jwks-url", "", "Set the JWKS URL to fetch public key for validation")
//
// // Postgres flags
// rootCommand.PersistentFlags().StringVarP(&postgres.Host, "postgres-host", "", postgres.Host, "Postgres host as IP address or name")
// rootCommand.PersistentFlags().StringVarP(&postgres.User, "postgres-user", "", postgres.User, "Postgres username")
// rootCommand.PersistentFlags().StringVarP(&postgres.Password, "postgres-password", "", postgres.Password, "Postgres password")
// rootCommand.PersistentFlags().StringVarP(&postgres.DBName, "postgres-dbname", "", postgres.DBName, "Postgres database name")
// rootCommand.PersistentFlags().StringVarP(&postgres.Opts, "postgres-opts", "", postgres.Opts, "Postgres database options")
// rootCommand.PersistentFlags().UintVarP(&postgres.Port, "postgres-port", "", postgres.Port, "Postgres port")
// rootCommand.PersistentFlags().Uint64VarP(&postgres.RetryCount, "postgres-retry_count", "", postgres.RetryCount, "Number of times to retry connecting to Postgres database before giving up")
// rootCommand.PersistentFlags().Uint64VarP(&postgres.RetryWait, "postgres-retry_wait", "", postgres.RetryWait, "Seconds to wait between retrying connection to Postgres")
// rootCommand.PersistentFlags().BoolVarP(&postgres.Insecure, "postgres-insecure", "", postgres.Insecure, "Don't enforce certificate authority for Postgres")

// TRC koanf wants to enforce a fairly strict struct via the delims--either the delims matter or they don't (flat)
// TRC I don't think there's a way to sorta combine them if we need to for backwards compatibility

// TRC broadly, I find this "the name of the flag/env/etc. determines the struct" model more confusing than "extract
// TRC this path into the struct field", though I suppose koanf does let you target by building a struct via k.String("key") values

// TRC AFAICT koanf is providing the viper bit (config and env parsing) only. it's _not_ providing most of cobra, only
// TRC the CLI arg reading. command compostion features (AddCommand, PreRun, etc. don't exist. it' _probably_
// TRC compatible with cobra--you'd basically set set no flags and would manually populate structs before passing them
// TRC to their consumer Run. I guess inheritance still sorta works through PersistentPreRun and such? but you're more
// TRC on your own.

// TRC the other two (alecthomas and urfave) do provide this. probably.

// TRC I think we can handle the common vars deal with a stock shared prefix we import and scan separately

type Config struct {
	// StateManagerURL and maybe StateManagerLockEnabled are good candidates for shared settings
	StateManagerURL         string `koanf:"smd"`
	StateManagerLockEnabled bool
	CredentialCacheDuration int
	MaxCompletedCount       int
	MaxCompletedTime        int
	JWKSURL                 string

	Postgres PostgresConfig
	Vault    VaultConfig
}

type VaultConfig struct {
	Enabled bool
	KeyPath string
}

type StateManagerConfig struct {
	URL         string
	LockEnabled string
}

type PostgresConfig struct {
	Host       string
	User       string
	Password   string
	Database   string
	Opts       string
	Port       uint
	RetryCount uint64
	RetryWait  uint64
	Insecure   bool
}

func main() {
	conf := koanf.New(".")
}

func lunk() {
	// TRC moved it
	// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
	// TRC The example has this as a global. I'd like to avoid this if possible for the usual inability to ensure init,
	// TRC is it required?
	var k = koanf.New(".")

	// TRC https://github.com/knadh/koanf?tab=readme-ov-file#unmarshalling-and-marshalling
	// TRC has this example but not whatever the heck does in to populate it.
	// TRC sadly https://github.com/knadh/koanf/tree/master/examples/unmarshal is just the go, and even the tests
	// TRC are basically empty. https://github.com/knadh/koanf/blob/master/tests/fs_test.go sorta has something, but
	// TRC not for structs
	// TRC however, https://github.com/knadh/koanf/blob/master/mock/mock.json apparently lines up with the example
	// TRC from that, yeah, it just aligns with the JSON nesting
	type childStruct struct {
		Time       string            `koanf:"time"`
		Type       string            `koanf:"type"`
		Empty      map[string]string `koanf:"empty"`
		GrandChild struct {
			Ids []int `koanf:"ids"`
			On  bool  `koanf:"on"`
		} `koanf:"grandchild"`
	}

	// TRC what's this "config" string end up used for. does koanf care about it?
	f := flag.NewFlagSet("config", flag.ContinueOnError)
	// TRC do you have to define this yourself? I guess this is technically a pflag thing. whatever
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	// TRC all the flag definitions are just regular old pflag
	f.String("time", "2020-01-01", "a time string")
	f.String("type", "xxx", "type of the app")
	f.Parse(os.Args[1:])

	// "time" and "type" may have been loaded from the config file, but
	// they can still be overridden with the values from the command line.
	// The bundled posflag.Provider takes a flagset from the spf13/pflag lib.
	// Passing the Koanf instance to posflag helps it deal with default command
	// line flag values that are not present in conf maps from previously loaded
	// providers.
	// TRC deleted the prior file layer, but yknow
	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	k.Load(env.Provider(".", env.Opt{
		// TRC shared prefix yay
		Prefix: "PCS_",
		// TRC yolo return type rather than matching keys to some struct field again
		TransformFunc: func(k, v string) (string, any) {
			// TRC ostensibly a TransformFunc is optional but it seems de facto required, you need it to strip prefixes
			// TRC and to change _ to another delimeter. is the delimiter handling just by convention with no parsing?
			// TRC maybe the Provider call inserts the base, but ultimately it's just a big string -> interface{} map?
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "PCS_")), "_", ".")

			// Transform the value into slices, if they contain spaces.
			// Eg: MYVAR_TAGS="foo bar baz" -> tags: ["foo", "bar", "baz"]
			// This is to demonstrate that string values can be transformed to any type
			// where necessary.
			if strings.Contains(v, " ") {
				return k, strings.Split(v, " ")
			}

			return k, v
		},
	}), nil)

	// TRC so you get params out by coercing arbitary strings out when you want them? is there no underlying struct?
	fmt.Println("time is = ", k.String("time"))
}

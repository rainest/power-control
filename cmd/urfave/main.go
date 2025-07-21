package main

// This packages shows how to use sflags with urfave/cli library.

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/urfave/cli/v3"
	"github.com/urfave/sflags"
	"github.com/urfave/sflags/gen/gcli"

	"github.com/OpenCHAMI/power-control/v2/internal/storage"
)

type httpConfig struct {
	Host    string ` desc:"HTTP host"`
	Port    int    `flag:"port p"`
	SSL     bool
	Timeout time.Duration
	Addr    *net.TCPAddr
}

type example struct {
	HTTP       httpConfig
	Regexp     *regexp.Regexp
	Count      sflags.Counter
	HiddenFlag string `flag:",hidden"`
}

const (
	defaultStateManagerURL   = "https://api-gw-service-nmn/apis/smd"
	defaultPort              = 28007
	defaultCompletedCount    = 20000 // Maximum number of completed records to keep (default 20k).
	defaultCompletedLifetime = 1440  // Time, in mins, to keep completed records (default 24 hours).
)

// TRC the root object for urfave/cli and sflags. automatic help output is in the same order as the struct fields.

type config struct {
	// StateManagerURL and maybe StateManagerLockEnabled are good candidates for shared settings
	CompletedCount    int `flag:"max-num-completed" desc:"Maximum number of completed records to keep."` // bikeshed breaking alternative completed-count
	CompletedLifetime int `flag:"expire-time-mins" desc:"Minutes to keep completed records."`            // bikeshed breaking alternative completed-lifetime
	// TRC this is a special case of the CamelCase -> --camel-case rule: all uppercase will be treated as one word, this defaults to --jwksurl
	JWKSURL string `flag:"jwks-url" desc:"Set the JWKS URL to fetch public key for validation"`
	Port    int    `flag:"port p" desc:"Service listen port"`
	// BREAKING runControl is gone. AFAICT it did nothing other than change a log message? it was supposed to only start the API, I think

	Vault        vaultConfig
	StateManager stateManagerConfig `flag:"smd"`

	Postgres postgresConfig
	Etcd     etcdConfig
}

type etcdConfig struct {
	DisableSizeChecks          bool `desc:"Disables checking object size before storing and doing message truncation and paging"`
	PageSize                   int  `desc:"The maximum number of records to put in each etcd entry"`
	MaxTransitionMessageLength int  `desc:"The maximum length of messages per task in a transition"`
	MaxObjectSize              int  `desc:"The maximum data size in bytes for objects in etcd"`
}

type vaultConfig struct {
	Disabled           bool   `desc:"Disable Vault for credentials"` // BREAKING the default is to have vault enabled, so this bool flips from --vault-enabled to --vault-disabled to actually work
	KeyPath            string `desc:"Key path for credentials in Vault"`
	CredentialLifetime int    `desc:"Seconds to cache credentials retrieved from Vault"` // BREAKING formerly cred-cache-duration
}

type stateManagerConfig struct {
	URL          string `desc:"SMD URL"`                    // BREAKING formerly sms-server
	LockDisabled bool   `desc:"Use hardware state Locking"` // BREAKING formerly hsmlock-enabled, renamed AND reversed because boolean
}

type postgresConfig struct {
	// TRC these are _IMPLICIT_: you only need to specify flags if they differ from the variable name.
	// TRC "CamelCase" automatically converts to --camel-case
	//Host       string `flag:"postgres-host"`
	//User       string `flag:"postgres-user"`
	//Password   string `flag:"postgres-password,hidden"`
	//Database   string `flag:"postgres-dbname"`
	//Opts       string `flag:"postgres-opts"`
	//Port       uint   `flag:"postgres-port"`
	//RetryCount uint64 `flag:"postgres-retry-count"` // originally using retry_count but w/e unreleased
	//RetryWait  uint64 `flag:"postgres-retry-wait"`  // ditto retry_wait
	//Insecure   bool   `flag:"postgres-insecure"`
	Host       string `desc:"Postgres hostname"`
	User       string `desc:"Postgres username"`
	Password   string `desc:"Postgres password"`
	Database   string `desc:"Postgres database name"` // BREAKING but not, dbname originally, but not yet released
	Opts       string `desc:"Postgres database options"`
	Port       uint   `desc:"Postgres port"`
	RetryCount uint64 `desc:"Number of times to retry connecting to Postgres database before giving up"` // originally using retry_count but w/e unreleased
	RetryWait  uint64 `desc:"Seconds to wait between retrying connection to Postgres"`                   // ditto retry_wait
	Insecure   bool   `desc:"Disable Postgres TLS validation"`
	// TRC this is wrong, flags are composed from their parent and automatically get its prefix, this manifests as --postgres-postgres-example
	ExampleSettingWrong string `flag:"postgres-example"`
	// TRC this is right
	ExampleSettingRight string `flag:"example"`
	// TRC this is just being stupid
	ExampleSettingWrongButRight string `flag:"~postgres-tricky-example"`
	// TRC this is overriding the parent prefix
	ExampleSettingOverride string       `flag:"~override-example"`
	Nested                 nestedConfig `flag:"nested"`
}

type nestedConfig struct {
	// TRC this gets both prefixes, so --postgres-nested-red and POSTGRES_NESTED_RED
	Red string `flag:"red"`
	// TRC overrides are absolute, this is --blue and BLUE, not POSTGRES_BLUE
	Blue string `flag:"~blue"`
	// TRC you probably shouldn't do this but it may be useful for refactoring struct structure
	Green string `flag:"~postgres-green"`
}

// TRC migrateConfig is an alt name since schemaConfig is the original config name, and I'm hacking around demonstrating multiple things at once
// TRC IDK if there's a better way to ensure subcommand flags share a common prefix (or if we just don't care). since
// TRC subcommands inherit their parents flags (at least when sflags handles them--you can set "Local: false" to not
// TRC propagate) you _don't_ want these in a struct with base config as anon and additional flags off a keyed field.
// TRC if you want a standard prefix, these need to have it in the var name or flag tag
type cmdMigrateConfig struct {
	SchemaStep       uint   `desc:"Migration step to apply"`
	SchemaForceStep  uint   `desc:""`
	SchemaFresh      bool   `desc:"Drop all tables and start fresh"`
	SchemaMigrations string `desc:"Directory for migration files"`
}

func DefaultConfig() *config {
	return &config{
		Port:              28007,
		CompletedCount:    20000,
		CompletedLifetime: 1440,
		Postgres: postgresConfig{
			User:     "pcsuser",
			Database: "pcs",
			Port:     5437,
		},
		Vault: vaultConfig{
			// TODO this default should change, purge the "HMS" nomenclature from ochami stuff.
			// IDK if we just have ochami-creds or what, we haven't really covered vault much before.
			KeyPath:            "secret/hms-creds",
			CredentialLifetime: 600,
		},
		StateManager: stateManagerConfig{
			URL: "https://api-gw-service-nmn/apis/smd",
		},
		Etcd: etcdConfig{
			PageSize:                   storage.DefaultEtcdPageSize,
			MaxTransitionMessageLength: storage.DefaultMaxMessageLen,
			MaxObjectSize:              storage.DefaultMaxEtcdObjectSize,
		},
	}
}

func DefaultMigrateConfig() *cmdMigrateConfig {
	return &cmdMigrateConfig{
		SchemaStep: 1,
	}
}

func main() {
	cfg := DefaultConfig()

	baseFlags, err := gcli.ParseV3(cfg, sflags.EnvPrefix("PCS_"))
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	mig := DefaultMigrateConfig()

	migrateFlags, err := gcli.ParseV3(mig, sflags.EnvPrefix("PCS_"))
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	cmd := &cli.Command{
		Name:        "power-control",
		Usage:       "Run the power-control service",
		Description: "OpenCHAMI power control service",
		Action: func(context.Context, *cli.Command) error {
			return nil
		},
		Flags: baseFlags,
		Commands: []*cli.Command{
			{
				Name:  "migrate-postgres",
				Usage: "Run Postgres database migrations",
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:    "schema-step",
						Usage:   "Migration step to apply",
						Sources: cli.EnvVars("PCS_SCHEMA_STEP"),
					},
					// This was originally int but seems like it shouldn't be. It's effectively the same data type
					// as a non-forced step.
					&cli.UintFlag{
						Name:    "schema-force-step",
						Usage:   "Force migration to a specific step",
						Sources: cli.EnvVars("PCS_SCHEMA_FORCE_STEP"),
					},
					&cli.BoolFlag{
						Name:    "schema-fresh",
						Usage:   "Drop all tables and start fresh",
						Sources: cli.EnvVars("PCS_SCHEMA_FRESH"),
					},
					&cli.StringFlag{
						Name:    "schema-migrations",
						Usage:   "Directory for migration files",
						Sources: cli.EnvVars("PCS_SCHEMA_MIGRATIONS"),
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					schema := &schemaConfig{}
					postgres := &storage.PostgresConfig{}
					migrateSchema(schema, postgres, nil)
					return nil
				},
			},
			{
				Name:  "alt-migrate",
				Usage: "Run Postgres database migrations",
				Flags: migrateFlags,
				Action: func(context.Context, *cli.Command) error {
					return nil
				},
			},
		},
	}

	//if err := cmd.Run(context.Background(), os.Args); err != nil {
	//	log.Fatal(err)
	//}

	//err = cmd.Run(context.Background(), []string{"power-control", "--help"})
	//if err != nil {
	//	fmt.Printf("err: %v", err)
	//}
	err = cmd.Run(context.Background(), []string{"power-control", "alt-migrate", "--help"})
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	err = cmd.Run(context.Background(), []string{
		"power-control",
		"alt-migrate",
		"--postgres-host", "postgres.example",
		"--postgres-password", "postgres.example",
		"--postgres-database", "pcsdb",
		// TRC booleans are simply present or not, so they MUST default to false. you CANNOT use --postgres-insecure false
		"--postgres-insecure",
		"--postgres-retry-count", "2000",
		"--postgres-example", "right",
		"--postgres-postgres-example", "wrong",
		"--postgres-tricky-example", "wronger",
		"--override-example", "trickier",
		"--postgres-nested-red", "red",
		"--blue", "blue",
		"--postgres-green", "green",
		"--schema-step", "5",
	})
	if err != nil {
		// TRC note that configuration parsing proceeds up to, but not past the first error. if you supply an argument
		// TRC that does not exist, you'll get every _prior_ argument into the struct, but nothing else after.
		// TRC this is mildly confusing since this example does not fatally exit and still prints the struct after errors,
		// TRC but in an actual app it'll presumably just fatal immediately.
		fmt.Printf("err: %v", err)
	}
	fmt.Printf("\ncfg: %s\n", spew.Sdump(cfg))
	fmt.Printf("\ncfg: %s\n", spew.Sdump(mig))
}

func exampleMain() {
	// TRC dunno whether to use consts or just stick things direct in here. original uses consts. ultimately we can't
	// TRC export them and there's not much reason to not inline them as such
	cfg := &config{
		Port:              28007,
		CompletedCount:    20000,
		CompletedLifetime: 1440,
		Postgres: postgresConfig{
			User:     "pcsuser",
			Database: "pcs",
			Port:     5437,
		},
		Vault: vaultConfig{
			// TODO this default should change, purge the "HMS" nomenclature from ochami stuff.
			// IDK if we just have ochami-creds or what, we haven't really covered vault much before.
			KeyPath:            "secret/hms-creds",
			CredentialLifetime: 600,
		},
		StateManager: stateManagerConfig{
			URL: "https://api-gw-service-nmn/apis/smd",
		},
		Etcd: etcdConfig{
			PageSize:                   storage.DefaultEtcdPageSize,
			MaxTransitionMessageLength: storage.DefaultMaxMessageLen,
			MaxObjectSize:              storage.DefaultMaxEtcdObjectSize,
		},
	}

	cmd := &cli.Command{
		Name:        "power-control",
		Usage:       "control power",
		Description: "OpenCHAMI power control service",
		Action: func(context.Context, *cli.Command) error {
			return nil
		},
	}

	err := gcli.ParseToV3(cfg, &cmd.Flags, sflags.EnvPrefix("PCS_"))
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

	// print usage
	err = cmd.Run(context.Background(), []string{"power-control", "--help"})
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	err = cmd.Run(context.Background(), []string{
		"power-control",
		"--postgres-host", "postgres.example",
		"--postgres-password", "postgres.example",
		"--postgres-database", "pcsdb",
		// TRC booleans are simply present or not, so they MUST default to false. you CANNOT use --postgres-insecure false
		"--postgres-insecure",
		"--postgres-retry-count", "2000",
		"--postgres-example", "right",
		"--postgres-postgres-example", "wrong",
		"--postgres-tricky-example", "wronger",
		"--override-example", "trickier",
		"--postgres-nested-red", "red",
		"--blue", "blue",
		"--postgres-green", "green",
	})
	if err != nil {
		// TRC note that configuration parsing proceeds up to, but not past the first error. if you supply an argument
		// TRC that does not exist, you'll get every _prior_ argument into the struct, but nothing else after.
		// TRC this is mildly confusing since this example does not fatally exit and still prints the struct after errors,
		// TRC but in an actual app it'll presumably just fatal immediately.
		fmt.Printf("err: %v", err)
	}
	fmt.Printf("\ncfg: %s\n", spew.Sdump(cfg))
}

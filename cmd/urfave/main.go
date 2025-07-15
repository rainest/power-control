package main

// This packages shows how to use sflags with urfave/cli library.

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/urfave/cli/v2"
	"github.com/urfave/sflags"
	"github.com/urfave/sflags/gen/gcli"
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
// "postgres-host", "", postgres.Host, "Postgres host as IP address or name")
// "postgres-user", "", postgres.User, "Postgres username")
// "postgres-password", "", postgres.Password, "Postgres password")
// "postgres-dbname", "", postgres.DBName, "Postgres database name")
// "postgres-opts", "", postgres.Opts, "Postgres database options")
// "postgres-port", "", postgres.Port, "Postgres port")
// "postgres-retry_count", "", postgres.RetryCount, "Number of times to retry connecting to Postgres database before giving up")
// "postgres-retry_wait", "", postgres.RetryWait, "Seconds to wait between retrying connection to Postgres")
// "postgres-insecure", "", postgres.Insecure, "Don't enforce certificate authority for Postgres")

type config struct {
	// StateManagerURL and maybe StateManagerLockEnabled are good candidates for shared settings
	CompletedCount    int    `flag:"max-num-completed"` // bikeshed alternative completed-count
	CompletedLifetime int    `flag:"expire-time-mins"`  // bikeshed alternative completed-lifetime
	JWKSURL           string `flag:"jwks-url"`
	Port              int    `flag:"port p"`

	Postgres     postgresConfig
	Vault        vaultConfig
	StateManager stateManagerConfig
}

type vaultConfig struct {
	Enabled                 bool   `flag:"vault-enabled"`
	KeyPath                 string `flag:"vault-keypath"`
	CredentialCacheDuration int    `flag:"vault-credential-lifetime"` // BREAKING formerly cred-cache-duration
}

type stateManagerConfig struct {
	URL         string `flag:"smd-url"`  // BREAKING formerly sms-server
	LockEnabled bool   `flag:"smd-lock"` // BREAKING formerly hsmlock-enabled
}

type postgresConfig struct {
	// TRC these
	//Host       string `flag:"postgres-host"`
	//User       string `flag:"postgres-user"`
	//Password   string `flag:"postgres-password,hidden"`
	//Database   string `flag:"postgres-dbname"`
	//Opts       string `flag:"postgres-opts"`
	//Port       uint   `flag:"postgres-port"`
	//RetryCount uint64 `flag:"postgres-retry-count"` // originally using retry_count but w/e unreleased
	//RetryWait  uint64 `flag:"postgres-retry-wait"`  // ditto retry_wait
	//Insecure   bool   `flag:"postgres-insecure"`
	Host                string
	User                string
	Password            string `flag:",hidden"`
	Database            string `flag:"dbname"`
	PathologicalExample string `flag:"postgres-pathological"`
	Opts                string
	Port                uint
	RetryCount          uint64 // originally using retry_count but w/e unreleased
	RetryWait           uint64 // ditto retry_wait
	Insecure            bool
}

func main() {
	// TRC dunno whether to use consts or just stick things direct in here. original uses consts. ultimately we can't
	// TRC export them and there's not much reason to not inline them as such
	cfg := &config{
		Port:              defaultPort,
		CompletedCount:    defaultCompletedCount,
		CompletedLifetime: defaultCompletedLifetime,
		Postgres: postgresConfig{
			User:                "pcsuser",
			Database:            "pcs",
			Port:                5437,
			PathologicalExample: "default",
		},
		Vault: vaultConfig{
			Enabled: true, // should this maybe be reversed? IIRC convention is that bools default to false
		},
		StateManager: stateManagerConfig{
			URL: defaultStateManagerURL,
		},
	}

	flags, err := gcli.Parse(cfg)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	cliApp := cli.NewApp()
	cliApp.Action = func(c *cli.Context) error {
		return nil
	}
	cliApp.Flags = flags
	// print usage
	err = cliApp.Run([]string{"cliApp", "--help"})
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	err = cliApp.Run([]string{
		"cliApp",
		"--postgres-host", "postgres.example",
		"--postgres-password", "postgres.example",
		"--postgres-dbname", "pcsdb",
		// TRC WRONG. this goes nowhere
		"--postgres-pathological", "poof",
		// TRC RIGHT.
		"--postgres-postgres-pathological", "hiiii",
	})
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	fmt.Printf("\ncfg: %s\n", spew.Sdump(cfg))
}

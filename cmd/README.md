# CLI library options

## Koanf

Provides argument/environment/file config parsing only. Does not provide a
command builder. It provides limited guardrails.

AFAIK no support for help generation or flag options (deprecated, hidden)
outside of pflag integration.

### Parsing

Koanf has a number of _Providers_, such as a YAML file or the environment, that
it can parse into the config map.

Providers coerce other formats into a nested `top.mid.keyA = 2`, `top.mid.keyB
= "foo"` shorthand for recursive `map[string]interface{}`. Downstream code
either unmarshals this into a struct or acceses keys directly with
`k.Int("top.mid.keyA")` and such.

#### Arguments

CLI parsing has a bad example from a world where all flags are one word:
https://github.com/knadh/koanf?tab=readme-ov-file#reading-from-command-line

TODO check if this automatically separates on dash.

#### Environment

https://github.com/knadh/koanf?tab=readme-ov-file#reading-from-command-line
doesn't mention the default behavior beyond filtering a prefix, unfortunately.

The example includes an optional transform function, which suggests you kinda
have to roll your own conversion from envvar name to koanf key. While this is
simple and will probably be the same "replace `_` with `.` each time, it feels
like that should be handled by default.

TODO is it? They may be confusingly demonstrating the default with code that
duplicates it without explaining as much.

#### Files

It's a YAML or things similar to YAML parser, objects are simple.

### Struct binding

https://github.com/knadh/koanf?tab=readme-ov-file#unmarshalling-and-marshalling
appears to enforce a fairly strict map between the raw configuration map and
the target struct. The flat approach allows you to circumvent this, but AFAICT
is all or nothing: there's no default unmarshal path, and you must specify
every target struct field's config map location.

This probably limits our flexibility in handling legacy structs or flag names.

### Other examples

The `ochami` CLI tool is sort of using it along with cobra, but not quite. For
example:

https://github.com/OpenCHAMI/ochami/blob/v0.3.4/cmd/bss-boot-script-get.go
defines flags via cobra only, and then executes functions that may pull config.

https://github.com/OpenCHAMI/ochami/blob/v0.3.4/cmd/lib.go#L334 calls 
https://github.com/OpenCHAMI/ochami/blob/v0.3.4/internal/config/config.go which
is handling koanf loads from a file. 

The cobra portion handles the xname list and other command arguments, whereas
the file config handles things like the service URL and user.

For services, I'd want a single point of configuration--they don't have the
same sort of natural separation that the tool has between per-invocation and
persistent settings.

## urfave/cli

Very batteries-included: parser, command builder, and assorted addons:
https://cli.urfave.org/v3/getting-started/

"playful and full of discovery" per the author. This isn't important, but I
found it an amusing description.

It's the clear winner by popularity, with 50k users and 20k stars, versus 2-3k
for koanf and kong. viper's at about 180k used by, though only about 30k stars.

### Struct binding

urfave maintains the separate https://github.com/urfave/sflags/ library to bind
structs. It uses struct tags with automatic inference from field names. It
allows overrides and supports hidden flags, deprecation, and description.

It will automatically populate help defaults with the existing field values
in the struct you pass to its parser.

Its example uses v2 instead of the latest v3 CLI library. Whoops!

It's quite nice for cutting down on boilerplate and keeping configuration close
to its destination, but apparently only supports a subset of the full `cli`
flag definition struct grammar. It's not clear if you can, for example, use
aliases with sflag, and the altsrc (file-based configuration) binding is
undocumented at best. It likely doesn't exist, as I don't see the altsrc
package in the sflags source.

#### File source approaches

sflags is effectively a code generator for `cli.Flag` structs. For example,
after parsing, I get:

```
(dlv) p cmd.Flags[0]
github.com/urfave/cli/v3.Flag(*github.com/urfave/cli/v3.FlagBase[github.com/urfave/cli/v3.Value,github.com/urfave/cli/v3.NoConfig,github.com/urfave/cli/v3.genericValue]) *{
	Name: "max-num-completed",
	Category: "",
	DefaultText: "",
	HideDefault: false,
	Usage: "Maximum number of completed records to keep.",
	Sources: github.com/urfave/cli/v3.ValueSourceChain {
		Chain: []github.com/urfave/cli/v3.ValueSource len: 1, cap: 1, [
			...,
		],},
	Required: false,
	Hidden: false,
	Local: false,
	Value: github.com/urfave/cli/v3.Value(*github.com/urfave/sflags/gen/gcli.value) *{
		v: github.com/urfave/sflags.Value(*github.com/urfave/sflags.intValue) ...,},
	Destination: *github.com/urfave/cli/v3.Value nil,
	Aliases: []string len: 0, cap: 0, nil,
	TakesFile: false,
	Action: nil,
	Config: github.com/urfave/cli/v3.NoConfig {},
	OnlyOnce: false,
	Validator: nil,
	ValidateDefaults: false,
	count: 0,
	hasBeenSet: false,
	applied: false,
	creator: github.com/urfave/cli/v3.genericValue {
		val: github.com/urfave/cli/v3.Value nil,},
	value: github.com/urfave/cli/v3.Value nil,}

(dlv) p cmd.Flags[0].Sources
github.com/urfave/cli/v3.ValueSourceChain {
	Chain: []github.com/urfave/cli/v3.ValueSource len: 1, cap: 1, [
		...,
	],}

(dlv) p cmd.Flags[0].Sources.Chain
[]github.com/urfave/cli/v3.ValueSource len: 1, cap: 1, [
	*github.com/urfave/cli/v3.envVarValueSource {
		Key: "PCS_MAX_NUM_COMPLETED",},
]
```

##### Adding generator support

The code for this in
https://github.com/urfave/sflags/blob/v0.4.1/gen/gcli/gcliv3.go#L33-L65 doesn't
currently support non-env sources, and adding additional ones after the fact is
probably annoying given we've got a slice, and no good way to infer the
original struct structure from its items.

As such adding file support would probably require hacking in something similar
to the [env parser](https://github.com/urfave/sflags/blob/v0.4.1/parser.go#L126-L167),
modifying the sflags.Flag type to have an additional Path field, updating the
[field parser](https://github.com/urfave/sflags/blob/v0.4.1/parser.go#L260) to
set it, and lastly modifying the `Sources` builder. This is a pretty tall
order, but not impossible.

The tag value would presumably follow the dotted string format in the altsrc
example, though IDK if we'd want a single or separate keys:

https://github.com/urfave/cli-altsrc

The keys would need to be distinct from the standard YAML and JSON tags.

##### YOLO don't care

Golang _already_ has perfectly fine support for building structs from YAML and
JSON. Nothing's really stopping us from just doing:

```
type config struct {
	Port    int    `flag:"port p" desc:"Service listen port" yaml:"port"`

	Vault        vaultConfig `yaml:"vault"`
}

type vaultConfig struct {
	Disabled           bool   `desc:"Disable Vault for credentials" yaml:"disabled"`
	KeyPath            string `desc:"Key path for credentials in Vault" yaml:"keyPath"`
	CredentialLifetime int    `desc:"Seconds to cache credentials retrieved from Vault" yaml:"credentialLifetime"`
}
```

and then either proving the unmarshaled result as either the default or a mergo
input. This does, however, annoyingly lose us the automatic inference.

Name inference is potentially annoying for case reasons anyway though.

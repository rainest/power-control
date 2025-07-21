# CLI library options

## Viper and Cobra

The de facto standard Golang CLI builder and argument manager, originally
developed for the Hugo static site generator: https://spf13.com/p/a-modern-cli-commander-for-go/

**Overall verdict**

The additional glue code required isn't "it just works!" but it's fairly
manageable. The major pro is that _some_ CSM code already uses this, saving us
some work. However, the CSM code wasn't necessarily using it well or
consistently, and definitely isn't using it portably.

**Details**

Commands are structs with various informative fields and lifecycle (run before,
run after, etc.) functions. Commands can have subcommands that inherit
lifecycle and args from their parents. Full CLIs are usually composed by adding
subcommands atop a root command:

```
rootCommand := createRootCommand(&pcs, &etcd, &postgres)
rootCommand.AddCommand(createPostgresInitCommand(&postgres, &schema))
```

The companion pflag package provides a richer, POSIX-compliant alternative to
the stdlib flags package. Flags target some variable, e.g.

```
cmd.Flags().BoolVar(&schema.fresh, "schema-fresh", schema.fresh, "Drop all tables and start fresh")
```
stores a `bool` in `schema.fresh`. While this commonly targets a big config
struct, it can target whatever.

The companion viper package provides support for envvars and file-based
configuration. This isn't handled by default; you need some glue code to
indicate which envvar matches which arg:

```
viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))

viper.SetEnvPrefix("APP") // binds APP_PORT, APP_ENV, APP_API_KEY, APP_API_SECRET
viper.AutomaticEnv()
```

The flag and envvar ideally match, but it's on you to ensure they do. Glue code
can automate this, but it's on you to provide it.

## Koanf

A configuration-only package that provides a collection of subpackages for
various sources. It allows accessing values through a JSONpath-esque syntax,
e.g. `koanf.String("path.to.thing")`.

**Overall verdict**

Too loosely-coupled for my taste. It's not very prescriptive on how various
sources are translated into its generic format, and I'd rather have my structs
up front than have them accessed through "trust me" path strings.

This seems like it's maybe useful for working with existing structured data
that you don't need all of, so you can pull arbitray keys out of a JSON or TOML
document without a full struct+tag definition, but I'm unsure what the use case
is if you control the whole config.

**Details**

This seems loosely designed around structured data formats like JSON as the
underlying principle. Their file example feels like a shortcut around the usual
tag structure:

```
// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
	// Load JSON config.
	if err := k.Load(file.Provider("mock/mock.json"), json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Load YAML config and merge into the previously loaded config (because we can).
	k.Load(file.Provider("mock/mock.yml"), yaml.Parser())

	fmt.Println("parent's name is = ", k.String("parent1.name"))
	fmt.Println("parent's ID is = ", k.Int("parent1.id"))
}
```

Though in their later unmarshal example they do have struct tags that
correspond to their string-delimited paths: https://github.com/knadh/koanf?tab=readme-ov-file#unmarshalling-and-marshalling

The envvar example has you rolling most of your own conversion to delimited
strings. This may be powerful, but I'd rather something prescriptive define a
format and then follow it: https://github.com/knadh/koanf?tab=readme-ov-file#reading-environment-variables

I'm a bit skeptical of the authors' arguments over viper--performance and size
aren't huge concerns for me (I'm not building tiny embedded apps, YOLO 500MB
binaries all the way) and prescriptivism with a few defined sources is
subjectively good: I don't really want an infinitely flexible format when I'm
probably just going to use flags and envvars anyway. It'd maybe be my choice if
I cared about config files only.

## Urfave

As someone who knows cobra first, I'd describe this as cobra but with viper
sorta integrated. It combines command building and flag management into one
package (albeit in different subpackages). It's somewhat less popular and most
discussion I can find has people moving off it to cobra.

**Overall verdict**

Being more tightly-coupled than cobra+viper+pflags is its major selling point
to me. The automatic inference from structs in sflags is intuitive and and
doesn't leave much to chance, but it's annoyingly cut off from some features in
the full definition.

You can sorta combine full definitions with sflags, but it's imperfect. If you
use full definition alone, it's arguably not much different from cobra/viper,
with the exception of defining flags and envvars in the same place without glue
code.

**Details**

I did demo code for this one, since it was my first alternate choice. See that.

## Kong

One struct to rule them all, with tags used for almost everything: https://github.com/alecthomas/kong?tab=readme-ov-file#supported-tags

**Overall verdict**

The duck-typed struct gives me the willies. Tags for everything feels like a
mess for all but the simplest CLIs, and I'm not keen on the lack of validation.
The terseness of the docs aren't helping things.

**Details**

Kong feels quite minimalist. Their base example is rather terse:

```
var CLI struct {
  Rm struct {
    Force     bool `help:"Force removal."`
    Recursive bool `help:"Recursively remove files."`

    Paths []string `arg:"" name:"path" help:"Paths to remove." type:"path"`
  } `cmd:"" help:"Remove files."`

  Ls struct {
    Paths []string `arg:"" optional:"" name:"path" help:"Paths to list." type:"path"`
  } `cmd:"" help:"List paths."`
}

func main() {
  ctx := kong.Parse(&CLI)
  switch ctx.Command() {
  case "rm <path>":
  case "ls":
  default:
    panic(ctx.Command())
  }
}
```

It's notable for not really distinguishing between commands and arguments: you
have one big struct whose fields may be subcommands _or_ flags _or_ args. By
default they're flags: https://github.com/alecthomas/kong?tab=readme-ov-file#flags

The properties of these are then determined by their tags, e.g. `default`
indicates the default subcommand, `hidden` marks flags _or_ commands that are
hidden.

Flags are generally automatically inferred from their field name, but support
aliases.

There's indication of support for envvars in the docs, but not much clear
documentation on how they're generated.

As best I can tell it populates the struct and then lets you do whatever in
your `Run()` definition--you reference the flag fields as needed.

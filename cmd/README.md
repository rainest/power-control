# CLI library options

## Koanf

Provides argument/environment/file config parsing only. Does not provide a
command builder.

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

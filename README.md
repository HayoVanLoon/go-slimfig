# Slimfig

A slim configuration library with a focus on ease of use for developers and
system operators.

## Quick Start

- Create a configuration file.

```
{
  "timeout_s": 90,
  "target": {
    "host": "localhost:8080",
    "schemes": ["http", "https"]
  }
}
```

- Pick a prefix, i.e. `XX`.
- Create `??_CONFIG` environment variable.

```shell
export XX_CONFIG=path/to/config.json
```

- Use your configuration.

```
import "github.com/HayoVanLoon/go-slimfig"

func main() {
    if err := slimfig.Load("XX"); err != nil {
        log.Fatal(err)
    }
    timeout := slimfig.Int("timeout_s", 60)
    host := slimfig.String("target.host", "")
    schemes := slimfig.String("target.schemes", "")
}
```

## Examples

Code examples can be found [here](./examples).

## Standard Patterns

### Overriding Configurations

This example continues the one from the quick start. Let's say we want to use a
different target host (for our production environment).

- Create second configuration file with values you want to override or add.

```
{
  "target": {
    "host": "www.example.com"
  },
  "auth": {
    "username": "admin",
    "password": "welcome01"
  }
}
```

- Update the `??_CONFIG` environment variable.

```shell
export XX_CONFIG=path/to/config.json,path/to/override.json
```

And that's it. Using this pattern allows you to:

- split up configuration files (for instance by feature)
- store a base or example configuration and override as desired
- separate sensitive values (like passwords) from non-sensitive ones.

Note that configuration loading is an all-or-nothing afair. Any error (a missing
file, invalid JSON, ...) will cause the loading to fail completely. The
configuration will remain empty.

### Using Google Secret Manager (or AWS)

Now instead of storing the override JSON in a secret volume, we want to store it
in Google Secret Manager.

- Create a new secret in Secret Manager, i.e. `my-override`.
- Store the override JSON data in it.
- Ensure the application has been authorised to access this secret.
- Update the `??_CONFIG` environment variable.

```shell
export XX_CONFIG=path/to/config.json,gcp-secretmanager://projects/123/secrets/my-override
```

- Customise the initialisation process to recognise Secret Manager.

```
import (
    "github.com/HayoVanLoon/go-slimfig"
    gcp "github.com/HayoVanLoon/go-slimfig/resolver/gcp/secret"
    "github.com/HayoVanLoon/go-slimfig/resolver/json"
)

func main() {
    secretManager, err := gcp.JSONResolver(ctx)
    if err != nil {
        log.Fatal(err)
    }
    slimfig.SetResolvers(secretManager, json.Resolver())

    // And continue as before ...
    if err := slimfig.Load("XX"); err != nil {
        ...
``` 

Provided authentication is properly set up, the default application credential
system will kick in and handle it. Alternatively, there is an option to provide
your own (fine-tuned) client.

For AWS Secrets Manager (note the extra 's'), the process is the same. Just load
another resolver and use `aws-secretsmanager://` (note the extra 's') in the
config variable. AWS Secret names do not contain paths, so you would just have
`aws-secretsmanager://my-override`.

## License

Copyright 2024 Hayo van Loon

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

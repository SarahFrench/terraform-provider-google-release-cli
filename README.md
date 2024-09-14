# google-provider-release-cli
A CLI to help making releases of the Google provider


## Configuring the tool

You need to create a file called `.tpg-cli-config.json` in your HOME directory:

```bash
touch $HOME/.tpg-cli-config.json
```

In that file you need to paste a JSON like below, which contains:
- magicModulesPath : the (absolute) path to where you have cloned the https://github.com/GoogleCloudPlatform/magic-modules repository
- googlePath : the (absolute) path to where you have cloned the https://github.com/hashicorp/terraform-provider-google repository
- googleBetaPath : the (absolute) path to where you have cloned the https://github.com/hashicorp/terraform-provider-google-beta repository
- remote : in your cloned copies of terraform-provider-google(-beta), the name of the "remote"  that corresponds to the official repo. If you're unsure, `cd` into those repos and run `git remote`.


```bash
vi  $HOME/.tpg-cli-config.json
```

```json
{
    "magicModulesPath": "/Users/Foobar/go/src/github.com/Foobar/magic-modules",
    "googlePath": "/Users/Foobar/go/src/github.com/Foobar/terraform-provider-google",
    "googleBetaPath": "/Users/Foobar/go/src/github.com/Foobar/terraform-provider-google-beta",
    "remote": "origin"
}
```
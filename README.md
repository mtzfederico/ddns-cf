# ddns-cf
This is a DDNS client for domains using Cloudflare written in Go.

## Usage
1. Clone the repo. `git clone https://github.com/mtzfederico/ddns-cf.git`
2. `cd ddns-cf`
3. Copy the sample config file and add your information (refer to table below). `cp sampleConfig.yaml config.yaml`
4. run `make buid` to compile the program. It will save a binary in the bin folder.
5. Use the included systemd timer files or a cronjob to run the binary.

To run the binary you have to add the `--config` parameter with the path to the config: `bin/ddns-cf --config config.yaml`

## Config Options

| Option            | Descrption                                                                                                                                                                 | Value Type | Required | Default Value |
|-------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------|----------|---------------|
| Domain            | The domain name to update                                                                                                                                                  | String     | yes      |               |
| SubDomainToUpdate | The subdomain of the Domain to update. If left empty, the Domain itself is used.                                                                                           | string     | no       |               |
| APIKey            | The Cloudflare API Key                                                                                                                                                     | string     | yes      |               |
| Email             | The Email address used in the Cloudflare account                                                                                                                           | string     | yes      |               |
| RecordTTL         | The TTL assigned to the domain in seconds. 1 sets it to cloudflare's automatic option.                                                                                     | int        | no       | Automatic     |
| IsProxied         | Use Cloudflare to proxy your traffic. Equivalent to enabling the cloud in Cloudflare.                                                                                      | bool       | no       | false         |
| DisableIPv4       | Disable checking and updating IPv4 and A Records                                                                                                                           | bool       | no       | false         |
| DisableIPv6       | Disable checking and updating IPv6 and AAAA Records                                                                                                                        | bool       | no       | false         |
| Verbose           | To print detailed log output.                                                                                                                                              | bool       | no       | false         |
| ScriptOnChange    | The path to a script or binary that gets executed when the IP address changes. The arguments are: the IP version ("v4" or "v6"), the old IP, the new IP, and the updated FQDN in that order. | string     | no       |               |
| LogFile           | The path to a file to save logs to.                                                                                                                                        | string     | no       |               |
| DebugLevel        | The level of details to log. The options from less detail to very detailed are: panic, fatal, error, warning, info, debug, and trace                                       | string     | no       | info (set by [logging library](https://github.com/sirupsen/logrus)) |
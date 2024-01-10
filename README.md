# sensu-ebs-monitoring

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The sensu-ebs-monitoring provides a diverse set of monitoring options for AWS EBS. 

## Usage examples

```
Usage:
  sensu-ebs-monitoring [flags]
  sensu-ebs-monitoring [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of this plugin

Flags:
      --region string                Sets the region of volume
      --volume-id string             Sets the volumeId
      --interval int                 Sets the interval in seconds for calcuating threshold status (default 60)
      --avg-readlatency string       Sets threshold in ms/op for read latency 
      --avg-writelatency string      Sets threshold in ms/op for write latency
      --max-iops string              Sets threshold for maximum IOPS
      --max-readops string           Sets maximum threshold for read operations
      --max-readthroughput string    Sets maximum threshold in KiB for read throughput
      --max-writeops string          Sets maximum threshold for write operations
      --max-writethroughput string   Sets maximum threshold in KiB for write throughput
      --min-iops string              Sets threshold for minimum IOPS
      --min-readops string           Sets minimum threshold for read operations
      --min-readthroughput string    Sets minimum threshold in KiB for read throughput
      --min-writeops string          Sets minimum threshold for write operations
      --min-writethroughput string   Sets minimum threshold in KiB for write throughput
      --nitro bool                   Sets whether the instance attached to ebs volume is nitro based (default false)
  -h, --help                         help for sensu-ebs-monitoring

Use "sensu-ebs-monitoring [command] --help" for more information about a command.
```

Note: It is important to specify the --region along with --volume-id followed by the desired check attribute.
The warning and critical values need to passed as command separated values. eg:- warning=<warningValue>,critical=<criticalValue>

### Example commands
Command to create a check which sets threshold for maximum iops :
```
sensu-ebs-monitoring --volume-id vol-0casdf98y9h988asd --region eu-west-1 --max-iops warning=8000,critical=10000 --interval 3600
```

Command to create a check which sets thresold for maximum read ops :
```
sensu-ebs-monitoring --volume-id vol-0casdf98y9h988asd --region eu-west-1 --max-readops critical=10000 --interval 3600
```

Command to create a check which sets thresold for average read latency for the volume attached to nitro based instance :
```
sensu-ebs-monitoring --volume-id vol-0casdf98y9h988asd --region eu-west-1 --max-readops warning=5,critical=10 --nitro true interval 60 
```




## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add Afaque-Anwar-Azad/sensu-ebs-monitoring
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/Afaque-Anwar-Azad/sensu-ebs-monitoring].

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: sensu-ebs-monitoring
  namespace: default
spec:
  command: sensu-ebs-monitoring --volume-id vol-0casdf98y9h988asd --region eu-west-1 --max-readops critical=10000
  subscriptions:
  - system
  runtime_assets:
  - Afaque-Anwar-Azad/sensu-ebs-monitoring
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-ebs-monitoring repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/Afaque-Anwar-Azad/sensu-ebs-monitoring/blob/master/.github/workflows/release.yml
[5]: https://github.com/Afaque-Anwar-Azad/sensu-ebs-monitoring/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[7]: https://github.com/sensu/check-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/

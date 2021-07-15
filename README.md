# cirocket

Rocket powered task runner to assist delivering ci build missions


## Status

This project is in its early alpha MVP stage.  It may be subject to change or redesign, but the hope is not to break any existing scripts using the tool.

Documentation is limited.  Current sources include this README and the sample config file generated by `cirocket init`

If you find and problem or want to contribute please see the contribution section below.

## Installation

To install either clone the repository or use the go tool chain to install.

```sh
git clone https://github.com/nehemming/cirocket.git
cd cirocket
go install

```
or

```sh
go install github.com/nehemming/cirocket@latest
```

# Basic usage

`cirocket` is a simple task runner application that uses a configuration file to specify the tasks.  

To create a sample config file, along with some basic documentation run:

```sh
cirocket init
```

This will create a `.cirocket.yml` file in the current working directory. 

Here is example [sample](https://github.com/nehemming/cirocket/blob/master/internal/cmd/initcoonfig.yml) 

Once edited the file can be run using

```sh
cirocket launch
```

Different mission files can be specified by adding to either command the `--config <filename>` flag.

## Contributing

We would welcome contributions to this project.  Please read our [CONTRIBUTION](https://github.com/nehemming/cirocket/blob/master/CONTRIBUTING.md) file for further details on how you can participate or report any issues.

## License

This software is licensed under the [Apache License](http://www.apache.org/licenses/). 
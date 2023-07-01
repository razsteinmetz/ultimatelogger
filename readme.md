# utlimatelogger - A versatile logger for golang

A library to allow logging to multiple outputs using logrus, with the following benefits:

* log to Telegram, Pushover, Console, Rotating files
* Configure different text formats for each output
* Configure different log levels for each output
* create an easy-to-use logger of logrus.Logger type
* singleton version, so can be init once and used everywhere without passing the logger around
* configuration can be done via a config file (toml)  or via code - and also by giving the config file path as a env variable
* standard, default configuration that allows the user to create a minimal config.toml file, or to change many options if needed
 


## Usage

There are two configuration files that are needed for the ultimate logger.   
One is the `defaultConfig.toml` - this defines default and most common parameters for each logging
output. Do not change the default config file. Use the specific config file to override values.  
The second is `ultimateLogger.toml` - this is the main configuration file. Here the user specify the parameters for each
service (api keys, logfile names etc.) and any other override of the defaults.

So the user need only create the ultimateLogger.toml and use:
On the first run the defaultConfig.toml will be created for you. 
you may use the sample ultimateLogger.toml file as a starting point.  
If the user does not create the config file `ultimatelogger.toml` a skeleton file will be created for you.


```go
package main
import "github.com/razstreinmetz/ultimateLogger"

func main() {
    log:=UltimateLogger()

	
	log.Info("This is an info message")
}
``` 

## Important Notice
The logging to telegram, pushover is done using go routines.  Thus if the program exist right after the logging
the messages might not be sent.  Add a short sleep before the exit to avoid it.
Panic and fatal messages will be waited a bit by default.

### TODO
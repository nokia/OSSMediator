# OpenNMSPlugin

The OpenNMSPlugin is a command line application that monitors the PM/FM directories where the mediator collector stores response and converts the PM/FM data to OpenNMS readable format and pushes it to OpenNMS data collection path.

### Prerequisites

OpenNMSPlugin is compatible with only Unix/Linux system.

### Project Structure

    .  
    ├── resources               # Resource files  
        └── resource.conf  
    ├── src                     # Source files  
    ├── Makefile  
    ├── Dockerfile  
    └── README.md  

### Installation steps

OpenNMSPlugin's binary should be built by running `make all` command.  
It will create binary named as `opennmsplugin` inside `bin` directory and package containing the binary and resource file, named as `OpenNMSPlugin.zip` inside `package` directory.  
  
Please follow below procedure to install OpenNMSPlugin.zip in your home directory:

````
$ mkdir mediator
$ cp OpenNMSPlugin.zip mediator/
$ cd mediator/
$ unzip OpenNMSPlugin.zip
````

OpenNMSPlugin directory structure after installation will be as shown below:

````
    .
    ├── OpenNMSPlugin.zip
    ├── bin
        └── opennmsplugin
    ├── log
        └── OpenNMSPlugin.log
    └── resources
        └── resource.conf
````

## Usage
Usage: ./opennmsplugin [options]  
Options:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-h, --help  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Output a usage message and exit.  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-conf_file  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Config file path (default "../resources/conf.json")  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-log_dir  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Log Directory (default "../log"), logs will be stored in OpenNMSPlugin.log file.  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-log_level  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Log Level (default 4), logger level in OpenNMSPlugin.log file. Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-v  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Prints OSSMediator's version  

## Configuration

OpenNMSPlugin reads all the collected PM/FM file to convert it to OpenNMS readable format and push it to OpenNMS data collection path.

* To convert PM/FM statistics, modify conf.json configuration file under the "resources" directory as shown in the example:

````json
{
    "users_conf": [
        {
            "source_dir": "<SOURCE DIRECTORY PATH>",
            "pm_config": {
                "destination_dir": "<PM DIRECTORY PATH>",
                "foreign_id": "<FOREIGN ID>"
            },
            "fm_config": {
                "source": "NDAC",
                "node_id": "1",
                "host": "127.0.0.1",
                "service": "NDAC",
                "destination_dir": "<FM DIRECTORY PATH>"
            }
        },
        {
            "source_dir": "<SOURCE DIRECTORY PATH>",
            "pm_config": {
                "destination_dir": "<PM DIRECTORY PATH>",
                "foreign_id": "<FOREIGN ID>"
            },
                "fm_config": {
                "source": "NDAC",
                "node_id": "1",
                "host": "127.0.0.1",
                "service": "NDAC",
                "destination_dir": "<FM DIRECTORY PATH>"
            }
        }
    ],
    "opennms_address": "127.0.0.1:5817",
    "cleanup_duration": 60
}
````

| Field                     | Type        | Description                                                               |
|---------------------------|-------------|---------------------------------------------------------------------------|
| users_conf                | [user_conf] | User's configurations.                                                    |
| source_dir                | string      | Base path of the respective user where PM/FM data is pushed by collector. |
| pm_config                 | object      | PM Config.                                                                |
| pm_config.destination_dir | string      | The path to push PM data after conversion.                                |
| pm_config.foreign_id      | string      | Unique ID of requisitions created in OpenNMS GUI.                         |
| fm_config                 | object      | FM Config.                                                                |
| fm_config.source          | string      | "Alarmd" is the constant value needed by OpenNMS.                         |
| fm_config.node_id         | string      | OpenNMS node for which alarm is generated.                                |
| fm_config.host            | string      | IP address of the OpenNMS server.                                         |
| fm_config.service         | string      | "NDAC" is the constant value needed by OpenNMS.                           |
| fm_config.destination_dir | string      | The path to push FM data after conversion.                                |
| opennms_address           | string      | IP address and port of OpenNMS server where FM data will be posted.       |
| cleanup_duration          | integer     | Duration at which OpenNMSPlugin will clean up already processed PM/FM files.   |

* To start opennmsplugin, go to the installed path of the mediator bin directory and start by calling the following command:

````
./opennmsplugin
````

* OpenNMSPlugin logs can be checked in $OpenNMSPlugin_HOME/log/OpenNMSPlugin.log file.

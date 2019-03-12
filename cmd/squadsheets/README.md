This project allows you to manage Squad Server Administrators easily through Google Sheets.

Create [Google API OAuth Client ID credentials](https://console.developers.google.com/apis/credentials) with the appropriate permissions. This will provide the relevant client_secret.json file for authentication you specify in the SquadSheets config.

Note: You will need to run SquadSheets in a cmd window before setting it up as a scheduled task to do the initial authentication. Instructions will be displayed through the command line.

## Usage 

```
NAME:
   squadsheets - syncs google sheets with admin data to squad admin files

USAGE:
   squadsheets [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value     configuration file (default: "squadsheets.cfg")
   --configdir value, -d value  Squad configuration directory
   --verbose                    verbose logging
   --log value, -l value        log file
   --whitelist, -w              ASC Whitelist
   --help, -h                   show help
   --version, -v                print the version

```

### Example

```
SquadSheets.exe -c C:\squad\squadsheets\squadsheets.cfg -d C:\squad\mysquadserver\squad\serverconfig -l C:\squad\squadsheets\squadsheets.log -w
```

Note: All sheets provide for, and will skip the first row assuming it's a header row
## Admin Roles Sheet
This sheet will provide admin role definitions for the admins file. NOTE: if a '__Whitelist__' role is not created, Squadsheets will create one with the ___reserve___ permission.

### Example

| Name | Capabilities |  
| ---- | ---- |  
| Whitelist | reserve |  
| ... | ... |  

## Admin Sheet
This sheet will provide admin assignments. You can have as many of these sheets as necessary as long as they are specified in the TOML array. This provides for some form of organization should you have different groups of admins.

### Example

| Name | Steam64 | Role | Notes | 
| ---- | ---- | ---- | ---- |
| Bob | 111111111111111 | SuperAdmin | Temporary admin | 
| ... | ... | ... | ... |

## Whitelist Sheet
This sheet will provide whitelist assignments. You can have as many of these sheets as necessary as long as they are specified in the TOML array. This provides for some form of organization should you have different groups of whitelisted users.

### Example

| Name | Steam64 | Notes | 
| ---- | ---- | ---- |
| Bob | 111111111111111 | Temporary whitelist | 
| ... | ... | ... |
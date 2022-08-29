# Changelog

## 29.08.2022, Version 1.6.4

- add schedule data source and resource

## 09.08.2022, Version 1.6.3

- fix parameter error send_notification on automation rules

## 28.07.2022, Version 1.6.2

- fix 429 error code problem if a lot of resources are applied

## 27.07.2022, Version 1.6.1

- fix [#19](https://github.com/iLert/terraform-provider-ilert/issues/19)

## 14.07.2022, Version 1.6.0

- add alert action data source and resource
- add service data source and resource
- add status page data source and resource
- add incident template data source and resource
- add automation rule resource
- deprecate connection data source and resource
- deprecate some legacy fields in resources

## 16.04.2022, Version 1.5.1

- fix connection resource trigger_types validation
- fix connection resource trigger_types docs
- fix email subject param in the connection resource
- fix user resource mobile and landline blocks
- fix all resources exists checks

## 14.04.2022, Version 1.5.0

- add timeout context to all resources
- replace custom validation functions
- minor bug fixes

## 19.01.2022, Version 1.4.5

- upgrade dependencies for github actions

## 19.01.2022, Version 1.4.4

- upgrade dependencies

## 18.01.2022, Version 1.4.3

- fix github actions

## 18.01.2022, Version 1.4.2

- fix ssl uptime monitor updates

## 18.01.2022, Version 1.4.1

- fix uptime monitor creation crash

## 18.01.2022, Version 1.4.0

- add more uptime monitor check params
- add new uptime monitor type: ssl

## 14.04.2021, Version 1.3.1

- add auto raise incidents prop to support hours

## 09.04.2021, Version 1.3.0

- add new alert source types
- add new connection types
- add new connector types

## 12.03.2021, Version 1.2.0

- add team data source and resource

## 16.11.2020, Version 1.1.3

- fix user language case

## 9.11.2020, Version 1.1.2

- add integration url to alert source data source and resource

## 8.11.2020, Version 1.1.1

- add jira alert source type

## 6.11.2020, Version 1.1.0

- add connection data source and resource
- add connector data source and resource
- remove type argument from escalation rule
- user standard user agent header for each request

## 22.10.2020, Version 1.0.0

- add alert source data source and resource
- add escalation policy data source and resource
- add schedule data source and resource
- add user data source and resource
- add uptime monitor data source and resource

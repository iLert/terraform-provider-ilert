# Changelog

- feature/alert-actions-v2 in [#60](https://github.com/iLert/terraform-provider-ilert/pull/60)

## 30.01.2024, Version 2.5.1

- fix/alert-source-support-hours-response in [#82](https://github.com/iLert/terraform-provider-ilert/pull/82)

## 12.01.2024, Version 2.5.0

- feature/add-telegram-alert-action-type in [#80](https://github.com/iLert/terraform-provider-ilert/pull/80)

## 05.01.2024, Version 2.4.1

- fix/status-page-theme-mode-field in [#79](https://github.com/iLert/terraform-provider-ilert/pull/79)

## 03.01.2024, Version 2.4.0

- fix/deprecate-uptime-monitors in [#74](https://github.com/iLert/terraform-provider-ilert/pull/74)
- feature/support-hours-resource in [#77](https://github.com/iLert/terraform-provider-ilert/pull/77)
  - fixes issue [#76](https://github.com/iLert/terraform-provider-ilert/issues/76)
- feature/status-page-layout-fields in [#78](https://github.com/iLert/terraform-provider-ilert/pull/78)

## 15.12.2023, Version 2.3.1

- fix/alert-action-delay-sec-restrictions in [#75](https://github.com/iLert/terraform-provider-ilert/pull/75)

## 12.12.2023, Version 2.3.0

- feature/improve-error-logging in [#71](https://github.com/iLert/terraform-provider-ilert/pull/71)
- feature/alert-source-link-priority-templates in [#73](https://github.com/iLert/terraform-provider-ilert/pull/73)

## 27.11.2023, Version 2.2.1

- fix/alert-action-delay-sec-validation in [#65](https://github.com/iLert/terraform-provider-ilert/pull/65)

## 13.11.2023, Version 2.2.0

- feature/alert-action-new-trigger-type-delaysec in [#63](https://github.com/iLert/terraform-provider-ilert/pull/63)

## 23.10.2023, Version 2.1.1

- fix/user-preferences-optional-contact in [#61](https://github.com/iLert/terraform-provider-ilert/pull/61)

## 09.10.2023, Version 2.1.0

- feature/escalation-policy-new-fields in [#57](https://github.com/iLert/terraform-provider-ilert/pull/57)
- feature/alert-source-new-fields in [#58](https://github.com/iLert/terraform-provider-ilert/pull/58)
- feature/alert-action-new-fields in [#59](https://github.com/iLert/terraform-provider-ilert/pull/59)

## 29.09.2023, Version 2.0.4

- fix/alert-action-jira-types in [#55](https://github.com/iLert/terraform-provider-ilert/pull/55)

## 02.05.2023, Version 2.0.3

- fix/status-page-missing-field-1 in [#53](https://github.com/iLert/terraform-provider-ilert/pull/53)

## 10.03.2023, Version 2.0.2

- update goreleaser in [#52](https://github.com/iLert/terraform-provider-ilert/pull/52)

## 10.03.2023, Version 2.0.1

- fix/go release 1.19 in [#51](https://github.com/iLert/terraform-provider-ilert/pull/51)

## 08.03.2023, Version 2.0.0 - API user preference migration: see [migration changes](https://docs.ilert.com/rest-api/api-version-history/api-user-preference-migration-2023#migrating-ilert-go-and-or-terraform) for a detailed migration guide

- feature/notification settings 2.0 in [#50](https://github.com/iLert/terraform-provider-ilert/pull/50)
  - remove notification settings from user resource
  - add user contacts
    - email
    - phone number
  - add user notification preferences
    - alert (alert creation)
    - duty (on-call)
    - subscription (subscriber to incident, service, status page)
    - update (alert update changes)
  - update all examples, add additional readme's
  - documentation overhaul

## 03.03.2023, Version 1.11.4

- fix/status-page-missing-field in [#49](https://github.com/iLert/terraform-provider-ilert/pull/49)
  - addresses issue [#48](https://github.com/iLert/terraform-provider-ilert/issues/48)

## 20.02.2023, Version 1.11.3

- fix/remove-deprecated-schema in [#46](https://github.com/iLert/terraform-provider-ilert/pull/46)
  - addresses issue [#45](https://github.com/iLert/terraform-provider-ilert/issues/45)

## 08.02.2023, Version 1.11.2

- fix/optional-username in [#44](https://github.com/iLert/terraform-provider-ilert/pull/44)

## 08.02.2023, Version 1.11.1

- fix/resource-schema-validation in [#43](https://github.com/iLert/terraform-provider-ilert/pull/43)
  - add nil checks to validation methods

## 18.01.2023, Version 1.11.0

- feature/metrics in [#42](https://github.com/iLert/terraform-provider-ilert/pull/42)
  - add metrics resource and data source
  - add metric data sources resource and data source

## 12.01.2023, Version 1.10.0

- fix/update-changelog-12_22 in [#39](https://github.com/iLert/terraform-provider-ilert/pull/39)
- fix/update-docs-deprecated-fields in [#40](https://github.com/iLert/terraform-provider-ilert/pull/40)
- feature/status-page-groups [#41](https://github.com/iLert/terraform-provider-ilert/pull/41)

## 28.11.2022, Version 1.9.1

- fix/alert action docs alert filter in [#38](https://github.com/iLert/terraform-provider-ilert/pull/38)

## 28.11.2022, Version 1.9.0

- fix/improve data source error messages in [#33](https://github.com/iLert/terraform-provider-ilert/pull/33)
- fix/alert action docs & iLert to ilert in [#34](https://github.com/iLert/terraform-provider-ilert/pull/34)
- feature/add alert filter field to alert action in [#35](https://github.com/iLert/terraform-provider-ilert/pull/35)
- feature/move automation rule status page ip filter in [#36](https://github.com/iLert/terraform-provider-ilert/pull/36)
- fix/add docu for automation rule type in [#37](https://github.com/iLert/terraform-provider-ilert/pull/37)

## 08.09.2022, Version 1.8.0

- add support for multiple responders in escalation rule (escalation policy)

## 07.09.2022, Version 1.7.3

- fix incident template + example
- fix schedule + example

## 05.09.2022, Version 1.7.2

- improve data source search for:
  - alert actions
  - alert sources
  - connectors
  - escalation policies
  - incident templates
  - schedules
  - services
  - status pages
  - teams
  - uptime monitors
  - users

## 01.09.2022, Version 1.7.1

- upgrade to go 1.19

## 01.09.2022, Version 1.7.0

- add schedule data source and resource
- fix issue that changes still happen if same resource is applied multiple times
  - fix service
  - fix status page
  - fix uptime monitor
  - fix user
  - fix team
- remove unnecessary field activated on statuspage

## 24.08.2022, Version 1.6.4

- fix [#22](https://github.com/iLert/terraform-provider-ilert/issues/22)

## 09.08.2022, Version 1.6.3

- fix parameter error send_notification on automation rules

## 28.07.2022, Version 1.6.2

- fix 429 error code problem if a lot of resources are applied

## 27.07.2022, Version 1.6.1

- fix [#19](https://github.com/iLert/terraform-provider-ilert/issues/19)

## 14.07.2022, Version 1.6.0 - API Version Update

### version renaming changes, see: https://docs.ilert.com/rest-api/api-version-history

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

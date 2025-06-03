# elastic-data

Repository of data that can be ingested into Elasticsearch using example data from elastic/integrations!

## TODO

- Allow to execute application without a TUI `elastic-data run`
- Show error messages to users
- Support more timestamp formats
  - snort needs more timestamps
  - crowdstrike uses  crowdstrike.metadata.eventCreationTime and need to adjust the json
- Go through JSON fields looking for common field names like the following and replace them in the template
  - username
  - hostname
  - time
- user can add their own domains, IPs, and emails
- user can provide their own log examples
- golangcilint
- write tests
- Generate a list of supported integrations and their datasets for users to easily look at
- Update README to explain what integrations/datasets are supported

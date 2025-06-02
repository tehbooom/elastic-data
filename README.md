# elastic-data

Repository of data that can be ingested into Elasticsearch using example data from elastic/integrations!

## TODO

- Show error messages to users
- Allow to execute application without a TUI `elastic-data run`
- Sending bytes stopping and starting restarts the count even if threshold is meet
- Update README to explain what integrations/datasets are supported
- Generate a list of supported integrations and their datasets for users to easily look at
- Support more timestamp formats
  - snort needs more timestamps
  - crowdstrike uses  crowdstrike.metadata.eventCreationTime and need to adjust the json
- Go through JSON fields looking for common field names like the following and replace them in the template
  - username
  - hostname
  - time
- Add search to integration selection
- golangcilint
- write tests
- user can add their own domains, IPs, and emails
- user can provide their own log examples
- when unselecting a integration it erases the dataset
- add support for preserving event original

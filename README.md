# elastic-data

Repository of data that can be ingested into Elasticsearch using example data from elastic/integrations!

## TODO

- Update help text
- Show error messages to users
- Allow to execute application without a TUI `elastic-data run`
- Sending bytes stopping and starting restarts the count even if threshold is meet
- Update README to explain what integrations/datasets are supported
- Generate a list of supported integrations and their datasets for users to easily look at
- Support multiline log files
  - support crowdstrike:falcon
  - Support more timestamp formats
- Display the integrations README when selecting datasets
- Add search to integration selection
- detect if dataset is logs or metrics (read in Type struct and its either metrics or logs)
- golangcilint
- write tests
- user can add their own domains, IPs, and emails
- user can provide their own log examples
- when unselecting a integration it erases the dataset

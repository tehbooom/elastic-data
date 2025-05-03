# elastic-data

Repository of data that can be ingested into Elasticsearch using example data from elastic/integrations!

User runs `elastic-data list-integrations --category=security``
User selects integrations: `elastic-data select-integrations apache nginx mysql`
> selected integrations are written to config file with defaults
User configures EPS/volume: `elastic-data configure-integration apache --eps=10`
> Updates the specified integration in the config file
User runs with config: `elastic-data run`

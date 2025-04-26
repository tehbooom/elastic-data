# elastic-data

Repository of data that can be ingested into Elasticsearch using example data from elastic/integrations!

1. Pull elastic/integrations
2. Select integrations
3. For each integration select datastreams to ingest (by default it selects all)
4. Installs the latest integration version into your stack
5. Configure the # of workers, EPS or Total events to ingest
6. Start the binary and run it in the background or wait for it to finish
7. OPTIONAL: Configure your own YAML in how it should be ingested. THis will create a template, datastream, and start indexing.

# elastic-data

Repository of data that can be ingested into Elasticsearch using example data from elastic/integrations!

1. Pull elastic/integrations
2. Select integrations
3. For each integration select datastreams to ingest (by default it selects all)
4. Installs the latest integration version into your stack
5. Configure the # of workers, EPS or Total events to ingest
6. Start the binary and run it in the background or wait for it to finish
7. OPTIONAL: Configure your own YAML in how it should be ingested. THis will create a template, datastream, and start indexing.


- Everything must be in a config file
- Test connection
- User can run the command and it will display each integration, the dataset, then the IO metrics


When a user runs the command it returns a loading screen to initialize, then it downloads the latest repo, then asks the user to confirm if they would like to start. Start? y/N


All integrations that are supported will be in the config file


Everything in a yaml file is going to be massive with 400 integrations and maybe 2 - 3 datasets per integration.

The best way is to do a tab view.

Loading screen no matter what to download the latest repo.

Then you are presented with a screen to select integrations

Tabs:

1. Integrations is going to be a list from left to right like a book maybe 3-4 columns. Users can search using item.List. They can select yes or no on an integration.

2. Datasets is where we select each data set and configure either EPS or total data in bytes to be sent.

3. Connection details this is debateable but we can have username, password, and APIkey, and the endpoints. The user should be able to press a key to test the connection. They can also update the authentcation and endpoints if needed.

4. Psuedo data is where the user can configure a list of names, IP addresses, ports, emails, etc to send the example data

5. Run tab is where we will see the metrics of IO to Elasticstack

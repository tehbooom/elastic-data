# elastic-data


<p>
    <a href="https://github.com/tehbooom/elastic-data/releases"><img src="https://img.shields.io/github/v/release/tehbooom/elastic-data.svg" alt="Latest Release"></a>
    <a href="https://github.com/tehbooom/elastic-data/blob/main/LICENSE"><img src="https://img.shields.io/github/license/tehbooom/elastic-data" alt="Latest Release"></a>
    <a href="https://goreportcard.com/report/github.com/tehbooom/elastic-data"><img src="https://goreportcard.com/badge/github.com/tehbooom/elastic-data" alt="GoDoc"></a>
    <a href="https://github.com/tehbooom/elastic-data/actions/workflows/lint.yml"><img src="https://github.com/tehbooom/elastic-data/actions/workflows/lint.yml/badge.svg" alt="Build Status"></a>
</p>


Repository of data that can be ingested into Elasticsearch using example data from [elastic/integrations](https://github.com/elastic/integrations)!

<p>
    <img src=".github/images/demo.gif" width="100%" alt="elastic-data example">
</p>

## Installation

Elastic-data can be installed by downloading the binary or installing from source.

### Binary

You can download the binary corresponding to your operating system from the releases page on GitHub.

Once downloaded you can run the binary from the command line:

```bash
tar -xzf elastic-data_Linux_x86_64.tar.gz
./elastic-data
```

### Build From Source

Ensure that you have a supported version of Go properly installed and setup. You can find the minimum required version of Go in the go.mod file.

You can then install the latest release globally by running:

```bash
go install github.com/tehbooom/elastic-data@latest
```

## Usage

> You must be authenticated with `git` before execution

1. Ensure you have a configuration file in `~/.config/elastic-data/config.yaml`

> If one does not exist on initial startup it will be created for you with some defaults

2. Select the integration(s) that you want to view data

3. For each integration(s) select the dataset(s) that you need and the following:

- Threshold
- Unit (eps or bytes)
- Preserve Original Event

4. Once saved go to the run tab and press `enter`

## Configuring

Below is the default configuration.

```yaml
connection:
  elasticsearch_endpoints:
      - https://localhost:9200
  kibana_endpoints:
      - https://localhost:5601
  password: changeme
  username: elastic
replacements:
  domains:
    - example.com
    - test.local
    - company.internal
  emails:
    - user@example.com
    - admin@company.com
    - noreply@test.local
  hostnames:
    - web-server-01
    - db-server
    - app-host
    - workstation-123
  ip_addresses:
    - 192.168.1.100
    - 10.0.0.50
    - 172.16.0.25
  usernames:
    - john.doe
    - admin
    - service_account
    - test_user
    - root
```

### Connection configuration

The default authentication method is username and password but you can also provide an API key.

```yaml
connection:
  api_key: abcd1234
```

For self signed certificates you can provide the certificate authority path as well as your own certificate and key if needed.

```yaml
connection:
  ca_cert: /path/to/ca
  cert: /path/to/cert
  key: /path/to/key
```

If you want to disable verification of certificates just add the unsafe flag

```yaml
connection:
  unsafe: true
```

### Replacement configuration

Sometimes the default replacements will not work for you. You can add or delete the default replacements to fit your needs.

```yaml
replacements:
  domains:
    - helloworld.io
  emails:
    - root@helloworld.io
  hostnames:
    - prod-db-01
  ip_addresses:
    - 8.8.8.8
  usernames:
    - supersecretuser
```


### Adding your own events

For some datasets you may want to use your own data as a template. You can do so by adding the following to the dataset

```yaml
integrations:
    nginx:
        datasets:
            access:
                enabled: true
                events:
                    - example.com localhost, localhost - - [29/May/2017:19:02:48 +0000] "PUT /test100 HTTP/1.1" 200 612 "-" "Mozilla/5.0 (Windows NT 6.1; rv:15.0) Gecko/20120716 Firefox/15.0a2" "-"
                preserve_original_event: true
                threshold: 10
                unit: eps
            error:
                enabled: false
                preserve_original_event: false
                threshold: 0
                unit: eps
        enabled: true
```

For JSON events add the events like so

```yaml
integrations:
    1password:
        datasets:
            audit_events:
                enabled: true
                events:
                    - '{"@timestamp": "2022-10-24T21:16:62.827288935Z", "message": "{\"uuid\": \"3UQOGUC7DVOCN4OZP2MDKHFLSG\",\"timestamp\": \"2022-10-24T21:16:52.827288935Z\",\"actor_uuid\": \"GLF6WUEKS5CSNDJ2OG6TCZD3M4\",\"actor_details\":{\"uuid\":\"GLF6WUEKS5CSNDJ2OG6TCZD3M4\", \"name\":\"Test user 34\", \"email\":\"test.actor@domain.com\"},\"action\": \"suspend\",\"object_type\": \"user\",\"object_uuid\":\"ZZZZZZZZ65AKHFETOUFO7NL4OM\",\"session\":{\"uuid\": \"ODOHXUYQCJBUJKRGZNNPBJURPE\",\"login_time\": \"2022-10-24T21:07:34.703106271Z\",\"device_uuid\":\"rqtd557fn2husnstp5nc66w2xa\",\"ip\":\"89.160.20.156\"},\"location\":{\"country\":\"Canada\",\"region\": \"Ontario\",\"city\": \"Toronto\",\"latitude\": 43.64,\"longitude\": -79.433}}"}'
                preserve_original_event: true
                threshold: 10
                unit: eps
            item_usages:
                enabled: false
                preserve_original_event: false
                threshold: 0
                unit: eps
            signin_attempts:
                enabled: false
                events: []
                preserve_original_event: false
                threshold: 1
                unit: eps
        enabled: true
```

For multiline events add the events like so

```yaml
integrations:
  snort:
    enabled: true
    datasets:
      log:
        enabled: true
        threshold: 5
        events:
          - |
            [**] [1:1000006:0] TCP connection [**]
            [Priority: 0] 
            09/04-21:42:42.860730 10.100.20.59:56012 -> 10.100.10.190:22
            TCP TTL:127 TOS:0x0 ID:53730 IpLen:20 DgmLen:108 DF
            ***AP*** Seq: 0x688E00E4  Ack: 0xBC730BB6  Win: 0x80B  TcpLen: 20
        preserve_original_event: true
        unit: eps
```

## Supported Integrations

Some integrations are not supported since their tests do no include example logs to generate data from.

<details>

<summary>List of integrations</summary>


- [1password](https://www.elastic.co/docs/reference/integrations/1password)

- [abnormal_security](https://www.elastic.co/docs/reference/integrations/abnormal_security)

- [activemq](https://www.elastic.co/docs/reference/integrations/activemq)

- [admin_by_request_epm](https://www.elastic.co/docs/reference/integrations/admin_by_request_epm)

- [amazon_security_lake](https://www.elastic.co/docs/reference/integrations/amazon_security_lake)

- [apache](https://www.elastic.co/docs/reference/integrations/apache)

- [apache_tomcat](https://www.elastic.co/docs/reference/integrations/apache_tomcat)

- [arista_ngfw](https://www.elastic.co/docs/reference/integrations/arista_ngfw)

- [armis](https://www.elastic.co/docs/reference/integrations/armis)

- [atlassian_bitbucket](https://www.elastic.co/docs/reference/integrations/atlassian_bitbucket)

- [atlassian_confluence](https://www.elastic.co/docs/reference/integrations/atlassian_confluence)

- [atlassian_jira](https://www.elastic.co/docs/reference/integrations/atlassian_jira)

- [auditd](https://www.elastic.co/docs/reference/integrations/auditd)

- [auditd_manager](https://www.elastic.co/docs/reference/integrations/auditd_manager)

- [auth0](https://www.elastic.co/docs/reference/integrations/auth0)

- [authentik](https://www.elastic.co/docs/reference/integrations/authentik)

- [aws](https://www.elastic.co/docs/reference/integrations/aws)

- [aws_bedrock](https://www.elastic.co/docs/reference/integrations/aws_bedrock)

- [aws_mq](https://www.elastic.co/docs/reference/integrations/aws_mq)

- [awsfirehose](https://www.elastic.co/docs/reference/integrations/awsfirehose)

- [azure](https://www.elastic.co/docs/reference/integrations/azure)

- [azure_app_service](https://www.elastic.co/docs/reference/integrations/azure_app_service)

- [azure_frontdoor](https://www.elastic.co/docs/reference/integrations/azure_frontdoor)

- [azure_functions](https://www.elastic.co/docs/reference/integrations/azure_functions)

- [azure_metrics](https://www.elastic.co/docs/reference/integrations/azure_metrics)

- [azure_network_watcher_nsg](https://www.elastic.co/docs/reference/integrations/azure_network_watcher_nsg)

- [azure_network_watcher_vnet](https://www.elastic.co/docs/reference/integrations/azure_network_watcher_vnet)

- [azure_openai](https://www.elastic.co/docs/reference/integrations/azure_openai)

- [barracuda](https://www.elastic.co/docs/reference/integrations/barracuda)

- [barracuda_cloudgen_firewall](https://www.elastic.co/docs/reference/integrations/barracuda_cloudgen_firewall)

- [bbot](https://www.elastic.co/docs/reference/integrations/bbot)

- [beelzebub](https://www.elastic.co/docs/reference/integrations/beelzebub)

- [beyondinsight_password_safe](https://www.elastic.co/docs/reference/integrations/beyondinsight_password_safe)

- [beyondtrust_pra](https://www.elastic.co/docs/reference/integrations/beyondtrust_pra)

- [bitdefender](https://www.elastic.co/docs/reference/integrations/bitdefender)

- [bitwarden](https://www.elastic.co/docs/reference/integrations/bitwarden)

- [blacklens](https://www.elastic.co/docs/reference/integrations/blacklens)

- [bluecoat](https://www.elastic.co/docs/reference/integrations/bluecoat)

- [box_events](https://www.elastic.co/docs/reference/integrations/box_events)

- [canva](https://www.elastic.co/docs/reference/integrations/canva)

- [carbon_black_cloud](https://www.elastic.co/docs/reference/integrations/carbon_black_cloud)

- [carbonblack_edr](https://www.elastic.co/docs/reference/integrations/carbonblack_edr)

- [cassandra](https://www.elastic.co/docs/reference/integrations/cassandra)

- [cef](https://www.elastic.co/docs/reference/integrations/cef)

- [ceph](https://www.elastic.co/docs/reference/integrations/ceph)

- [checkpoint](https://www.elastic.co/docs/reference/integrations/checkpoint)

- [checkpoint_email](https://www.elastic.co/docs/reference/integrations/checkpoint_email)

- [checkpoint_harmony_endpoint](https://www.elastic.co/docs/reference/integrations/checkpoint_harmony_endpoint)

- [cisa_kevs](https://www.elastic.co/docs/reference/integrations/cisa_kevs)

- [cisco_aironet](https://www.elastic.co/docs/reference/integrations/cisco_aironet)

- [cisco_asa](https://www.elastic.co/docs/reference/integrations/cisco_asa)

- [cisco_duo](https://www.elastic.co/docs/reference/integrations/cisco_duo)

- [cisco_ftd](https://www.elastic.co/docs/reference/integrations/cisco_ftd)

- [cisco_ios](https://www.elastic.co/docs/reference/integrations/cisco_ios)

- [cisco_ise](https://www.elastic.co/docs/reference/integrations/cisco_ise)

- [cisco_meraki](https://www.elastic.co/docs/reference/integrations/cisco_meraki)

- [cisco_meraki_metrics](https://www.elastic.co/docs/reference/integrations/cisco_meraki_metrics)

- [cisco_nexus](https://www.elastic.co/docs/reference/integrations/cisco_nexus)

- [cisco_secure_email_gateway](https://www.elastic.co/docs/reference/integrations/cisco_secure_email_gateway)

- [cisco_secure_endpoint](https://www.elastic.co/docs/reference/integrations/cisco_secure_endpoint)

- [cisco_umbrella](https://www.elastic.co/docs/reference/integrations/cisco_umbrella)

- [citrix_adc](https://www.elastic.co/docs/reference/integrations/citrix_adc)

- [citrix_waf](https://www.elastic.co/docs/reference/integrations/citrix_waf)

- [claroty_ctd](https://www.elastic.co/docs/reference/integrations/claroty_ctd)

- [claroty_xdome](https://www.elastic.co/docs/reference/integrations/claroty_xdome)

- [cloud_security_posture](https://www.elastic.co/docs/reference/integrations/cloud_security_posture)

- [cloudflare](https://www.elastic.co/docs/reference/integrations/cloudflare)

- [cloudflare_logpush](https://www.elastic.co/docs/reference/integrations/cloudflare_logpush)

- [coredns](https://www.elastic.co/docs/reference/integrations/coredns)

- [couchbase](https://www.elastic.co/docs/reference/integrations/couchbase)

- [couchdb](https://www.elastic.co/docs/reference/integrations/couchdb)

- [crowdstrike](https://www.elastic.co/docs/reference/integrations/crowdstrike)

- [cyberark_epm](https://www.elastic.co/docs/reference/integrations/cyberark_epm)

- [cyberark_pta](https://www.elastic.co/docs/reference/integrations/cyberark_pta)

- [cyberarkpas](https://www.elastic.co/docs/reference/integrations/cyberarkpas)

- [cybereason](https://www.elastic.co/docs/reference/integrations/cybereason)

- [cylance](https://www.elastic.co/docs/reference/integrations/cylance)

- [darktrace](https://www.elastic.co/docs/reference/integrations/darktrace)

- [digital_guardian](https://www.elastic.co/docs/reference/integrations/digital_guardian)

- [elastic_package_registry](https://www.elastic.co/docs/reference/integrations/elastic_package_registry)

- [elasticsearch](https://www.elastic.co/docs/reference/integrations/elasticsearch)

- [endace](https://www.elastic.co/docs/reference/integrations/endace)

- [entityanalytics_ad](https://www.elastic.co/docs/reference/integrations/entityanalytics_ad)

- [entityanalytics_entra_id](https://www.elastic.co/docs/reference/integrations/entityanalytics_entra_id)

- [entityanalytics_okta](https://www.elastic.co/docs/reference/integrations/entityanalytics_okta)

- [envoyproxy](https://www.elastic.co/docs/reference/integrations/envoyproxy)

- [eset_protect](https://www.elastic.co/docs/reference/integrations/eset_protect)

- [f5_bigip](https://www.elastic.co/docs/reference/integrations/f5_bigip)

- [falco](https://www.elastic.co/docs/reference/integrations/falco)

- [fireeye](https://www.elastic.co/docs/reference/integrations/fireeye)

- [first_epss](https://www.elastic.co/docs/reference/integrations/first_epss)

- [forcepoint_web](https://www.elastic.co/docs/reference/integrations/forcepoint_web)

- [forgerock](https://www.elastic.co/docs/reference/integrations/forgerock)

- [fortinet_forticlient](https://www.elastic.co/docs/reference/integrations/fortinet_forticlient)

- [fortinet_fortiedr](https://www.elastic.co/docs/reference/integrations/fortinet_fortiedr)

- [fortinet_fortigate](https://www.elastic.co/docs/reference/integrations/fortinet_fortigate)

- [fortinet_fortimail](https://www.elastic.co/docs/reference/integrations/fortinet_fortimail)

- [fortinet_fortimanager](https://www.elastic.co/docs/reference/integrations/fortinet_fortimanager)

- [fortinet_fortiproxy](https://www.elastic.co/docs/reference/integrations/fortinet_fortiproxy)

- [gcp](https://www.elastic.co/docs/reference/integrations/gcp)

- [gcp_vertexai](https://www.elastic.co/docs/reference/integrations/gcp_vertexai)

- [gigamon](https://www.elastic.co/docs/reference/integrations/gigamon)

- [github](https://www.elastic.co/docs/reference/integrations/github)

- [gitlab](https://www.elastic.co/docs/reference/integrations/gitlab)

- [goflow2](https://www.elastic.co/docs/reference/integrations/goflow2)

- [golang](https://www.elastic.co/docs/reference/integrations/golang)

- [google_scc](https://www.elastic.co/docs/reference/integrations/google_scc)

- [google_secops](https://www.elastic.co/docs/reference/integrations/google_secops)

- [google_workspace](https://www.elastic.co/docs/reference/integrations/google_workspace)

- [hadoop](https://www.elastic.co/docs/reference/integrations/hadoop)

- [haproxy](https://www.elastic.co/docs/reference/integrations/haproxy)

- [hashicorp_vault](https://www.elastic.co/docs/reference/integrations/hashicorp_vault)

- [hid_bravura_monitor](https://www.elastic.co/docs/reference/integrations/hid_bravura_monitor)

- [hpe_aruba_cx](https://www.elastic.co/docs/reference/integrations/hpe_aruba_cx)

- [ibmmq](https://www.elastic.co/docs/reference/integrations/ibmmq)

- [iis](https://www.elastic.co/docs/reference/integrations/iis)

- [imperva](https://www.elastic.co/docs/reference/integrations/imperva)

- [imperva_cloud_waf](https://www.elastic.co/docs/reference/integrations/imperva_cloud_waf)

- [infoblox_bloxone_ddi](https://www.elastic.co/docs/reference/integrations/infoblox_bloxone_ddi)

- [infoblox_nios](https://www.elastic.co/docs/reference/integrations/infoblox_nios)

- [iptables](https://www.elastic.co/docs/reference/integrations/iptables)

- [istio](https://www.elastic.co/docs/reference/integrations/istio)

- [jamf_compliance_reporter](https://www.elastic.co/docs/reference/integrations/jamf_compliance_reporter)

- [jamf_pro](https://www.elastic.co/docs/reference/integrations/jamf_pro)

- [jamf_protect](https://www.elastic.co/docs/reference/integrations/jamf_protect)

- [jumpcloud](https://www.elastic.co/docs/reference/integrations/jumpcloud)

- [juniper_junos](https://www.elastic.co/docs/reference/integrations/juniper_junos)

- [juniper_netscreen](https://www.elastic.co/docs/reference/integrations/juniper_netscreen)

- [juniper_srx](https://www.elastic.co/docs/reference/integrations/juniper_srx)

- [kafka](https://www.elastic.co/docs/reference/integrations/kafka)

- [keycloak](https://www.elastic.co/docs/reference/integrations/keycloak)

- [kibana](https://www.elastic.co/docs/reference/integrations/kibana)

- [kubernetes](https://www.elastic.co/docs/reference/integrations/kubernetes)

- [lastpass](https://www.elastic.co/docs/reference/integrations/lastpass)

- [logstash](https://www.elastic.co/docs/reference/integrations/logstash)

- [lumos](https://www.elastic.co/docs/reference/integrations/lumos)

- [lyve_cloud](https://www.elastic.co/docs/reference/integrations/lyve_cloud)

- [m365_defender](https://www.elastic.co/docs/reference/integrations/m365_defender)

- [mattermost](https://www.elastic.co/docs/reference/integrations/mattermost)

- [menlo](https://www.elastic.co/docs/reference/integrations/menlo)

- [microsoft_defender_cloud](https://www.elastic.co/docs/reference/integrations/microsoft_defender_cloud)

- [microsoft_defender_endpoint](https://www.elastic.co/docs/reference/integrations/microsoft_defender_endpoint)

- [microsoft_dhcp](https://www.elastic.co/docs/reference/integrations/microsoft_dhcp)

- [microsoft_dnsserver](https://www.elastic.co/docs/reference/integrations/microsoft_dnsserver)

- [microsoft_exchange_online_message_trace](https://www.elastic.co/docs/reference/integrations/microsoft_exchange_online_message_trace)

- [microsoft_exchange_server](https://www.elastic.co/docs/reference/integrations/microsoft_exchange_server)

- [microsoft_sentinel](https://www.elastic.co/docs/reference/integrations/microsoft_sentinel)

- [microsoft_sqlserver](https://www.elastic.co/docs/reference/integrations/microsoft_sqlserver)

- [mimecast](https://www.elastic.co/docs/reference/integrations/mimecast)

- [miniflux](https://www.elastic.co/docs/reference/integrations/miniflux)

- [modsecurity](https://www.elastic.co/docs/reference/integrations/modsecurity)

- [mongodb](https://www.elastic.co/docs/reference/integrations/mongodb)

- [mongodb_atlas](https://www.elastic.co/docs/reference/integrations/mongodb_atlas)

- [mysql](https://www.elastic.co/docs/reference/integrations/mysql)

- [mysql_enterprise](https://www.elastic.co/docs/reference/integrations/mysql_enterprise)

- [nagios_xi](https://www.elastic.co/docs/reference/integrations/nagios_xi)

- [nats](https://www.elastic.co/docs/reference/integrations/nats)

- [netflow](https://www.elastic.co/docs/reference/integrations/netflow)

- [netscout](https://www.elastic.co/docs/reference/integrations/netscout)

- [netskope](https://www.elastic.co/docs/reference/integrations/netskope)

- [network_traffic](https://www.elastic.co/docs/reference/integrations/network_traffic)

- [nginx](https://www.elastic.co/docs/reference/integrations/nginx)

- [nginx_ingress_controller](https://www.elastic.co/docs/reference/integrations/nginx_ingress_controller)

- [o365](https://www.elastic.co/docs/reference/integrations/o365)

- [o365_metrics](https://www.elastic.co/docs/reference/integrations/o365_metrics)

- [okta](https://www.elastic.co/docs/reference/integrations/okta)

- [opencanary](https://www.elastic.co/docs/reference/integrations/opencanary)

- [oracle](https://www.elastic.co/docs/reference/integrations/oracle)

- [oracle_weblogic](https://www.elastic.co/docs/reference/integrations/oracle_weblogic)

- [osquery](https://www.elastic.co/docs/reference/integrations/osquery)

- [panw](https://www.elastic.co/docs/reference/integrations/panw)

- [panw_cortex_xdr](https://www.elastic.co/docs/reference/integrations/panw_cortex_xdr)

- [panw_metrics](https://www.elastic.co/docs/reference/integrations/panw_metrics)

- [pfsense](https://www.elastic.co/docs/reference/integrations/pfsense)

- [php_fpm](https://www.elastic.co/docs/reference/integrations/php_fpm)

- [ping_federate](https://www.elastic.co/docs/reference/integrations/ping_federate)

- [ping_one](https://www.elastic.co/docs/reference/integrations/ping_one)

- [platform_observability](https://www.elastic.co/docs/reference/integrations/platform_observability)

- [postgresql](https://www.elastic.co/docs/reference/integrations/postgresql)

- [pps](https://www.elastic.co/docs/reference/integrations/pps)

- [prisma_access](https://www.elastic.co/docs/reference/integrations/prisma_access)

- [prisma_cloud](https://www.elastic.co/docs/reference/integrations/prisma_cloud)

- [proofpoint_itm](https://www.elastic.co/docs/reference/integrations/proofpoint_itm)

- [proofpoint_on_demand](https://www.elastic.co/docs/reference/integrations/proofpoint_on_demand)

- [proofpoint_tap](https://www.elastic.co/docs/reference/integrations/proofpoint_tap)

- [proxysg](https://www.elastic.co/docs/reference/integrations/proxysg)

- [pulse_connect_secure](https://www.elastic.co/docs/reference/integrations/pulse_connect_secure)

- [qnap_nas](https://www.elastic.co/docs/reference/integrations/qnap_nas)

- [qualys_vmdr](https://www.elastic.co/docs/reference/integrations/qualys_vmdr)

- [rabbitmq](https://www.elastic.co/docs/reference/integrations/rabbitmq)

- [rapid7_insightvm](https://www.elastic.co/docs/reference/integrations/rapid7_insightvm)

- [redis](https://www.elastic.co/docs/reference/integrations/redis)

- [rubrik](https://www.elastic.co/docs/reference/integrations/rubrik)

- [sailpoint_identity_sc](https://www.elastic.co/docs/reference/integrations/sailpoint_identity_sc)

- [salesforce](https://www.elastic.co/docs/reference/integrations/salesforce)

- [santa](https://www.elastic.co/docs/reference/integrations/santa)

- [sentinel_one](https://www.elastic.co/docs/reference/integrations/sentinel_one)

- [sentinel_one_cloud_funnel](https://www.elastic.co/docs/reference/integrations/sentinel_one_cloud_funnel)

- [servicenow](https://www.elastic.co/docs/reference/integrations/servicenow)

- [slack](https://www.elastic.co/docs/reference/integrations/slack)

- [snort](https://www.elastic.co/docs/reference/integrations/snort)

- [snyk](https://www.elastic.co/docs/reference/integrations/snyk)

- [sonicwall_firewall](https://www.elastic.co/docs/reference/integrations/sonicwall_firewall)

- [sophos](https://www.elastic.co/docs/reference/integrations/sophos)

- [sophos_central](https://www.elastic.co/docs/reference/integrations/sophos_central)

- [splunk](https://www.elastic.co/docs/reference/integrations/splunk)

- [spring_boot](https://www.elastic.co/docs/reference/integrations/spring_boot)

- [spycloud](https://www.elastic.co/docs/reference/integrations/spycloud)

- [squid](https://www.elastic.co/docs/reference/integrations/squid)

- [stan](https://www.elastic.co/docs/reference/integrations/stan)

- [stormshield](https://www.elastic.co/docs/reference/integrations/stormshield)

- [sublime_security](https://www.elastic.co/docs/reference/integrations/sublime_security)

- [suricata](https://www.elastic.co/docs/reference/integrations/suricata)

- [swimlane](https://www.elastic.co/docs/reference/integrations/swimlane)

- [symantec_endpoint](https://www.elastic.co/docs/reference/integrations/symantec_endpoint)

- [symantec_endpoint_security](https://www.elastic.co/docs/reference/integrations/symantec_endpoint_security)

- [sysdig](https://www.elastic.co/docs/reference/integrations/sysdig)

- [syslog_router](https://www.elastic.co/docs/reference/integrations/syslog_router)

- [sysmon_linux](https://www.elastic.co/docs/reference/integrations/sysmon_linux)

- [system](https://www.elastic.co/docs/reference/integrations/system)

- [system_audit](https://www.elastic.co/docs/reference/integrations/system_audit)

- [tanium](https://www.elastic.co/docs/reference/integrations/tanium)

- [teleport](https://www.elastic.co/docs/reference/integrations/teleport)

- [tenable_io](https://www.elastic.co/docs/reference/integrations/tenable_io)

- [tenable_ot_security](https://www.elastic.co/docs/reference/integrations/tenable_ot_security)

- [tenable_sc](https://www.elastic.co/docs/reference/integrations/tenable_sc)

- [tencent_cloud](https://www.elastic.co/docs/reference/integrations/tencent_cloud)

- [tetragon](https://www.elastic.co/docs/reference/integrations/tetragon)

- [thycotic_ss](https://www.elastic.co/docs/reference/integrations/thycotic_ss)

- [ti_abusech](https://www.elastic.co/docs/reference/integrations/ti_abusech)

- [ti_anomali](https://www.elastic.co/docs/reference/integrations/ti_anomali)

- [ti_cif3](https://www.elastic.co/docs/reference/integrations/ti_cif3)

- [ti_crowdstrike](https://www.elastic.co/docs/reference/integrations/ti_crowdstrike)

- [ti_custom](https://www.elastic.co/docs/reference/integrations/ti_custom)

- [ti_cybersixgill](https://www.elastic.co/docs/reference/integrations/ti_cybersixgill)

- [ti_domaintools](https://www.elastic.co/docs/reference/integrations/ti_domaintools)

- [ti_eclecticiq](https://www.elastic.co/docs/reference/integrations/ti_eclecticiq)

- [ti_eset](https://www.elastic.co/docs/reference/integrations/ti_eset)

- [ti_maltiverse](https://www.elastic.co/docs/reference/integrations/ti_maltiverse)

- [ti_mandiant_advantage](https://www.elastic.co/docs/reference/integrations/ti_mandiant_advantage)

- [ti_misp](https://www.elastic.co/docs/reference/integrations/ti_misp)

- [ti_opencti](https://www.elastic.co/docs/reference/integrations/ti_opencti)

- [ti_otx](https://www.elastic.co/docs/reference/integrations/ti_otx)

- [ti_rapid7_threat_command](https://www.elastic.co/docs/reference/integrations/ti_rapid7_threat_command)

- [ti_recordedfuture](https://www.elastic.co/docs/reference/integrations/ti_recordedfuture)

- [ti_threatconnect](https://www.elastic.co/docs/reference/integrations/ti_threatconnect)

- [ti_threatq](https://www.elastic.co/docs/reference/integrations/ti_threatq)

- [tines](https://www.elastic.co/docs/reference/integrations/tines)

- [tomcat](https://www.elastic.co/docs/reference/integrations/tomcat)

- [traefik](https://www.elastic.co/docs/reference/integrations/traefik)

- [trellix_edr_cloud](https://www.elastic.co/docs/reference/integrations/trellix_edr_cloud)

- [trellix_epo_cloud](https://www.elastic.co/docs/reference/integrations/trellix_epo_cloud)

- [trend_micro_vision_one](https://www.elastic.co/docs/reference/integrations/trend_micro_vision_one)

- [trendmicro](https://www.elastic.co/docs/reference/integrations/trendmicro)

- [tychon](https://www.elastic.co/docs/reference/integrations/tychon)

- [varonis](https://www.elastic.co/docs/reference/integrations/varonis)

- [vectra_detect](https://www.elastic.co/docs/reference/integrations/vectra_detect)

- [vectra_rux](https://www.elastic.co/docs/reference/integrations/vectra_rux)

- [vsphere](https://www.elastic.co/docs/reference/integrations/vsphere)

- [watchguard_firebox](https://www.elastic.co/docs/reference/integrations/watchguard_firebox)

- [websphere_application_server](https://www.elastic.co/docs/reference/integrations/websphere_application_server)

- [windows](https://www.elastic.co/docs/reference/integrations/windows)

- [wiz](https://www.elastic.co/docs/reference/integrations/wiz)

- [zeek](https://www.elastic.co/docs/reference/integrations/zeek)

- [zerofox](https://www.elastic.co/docs/reference/integrations/zerofox)

- [zeronetworks](https://www.elastic.co/docs/reference/integrations/zeronetworks)

- [zoom](https://www.elastic.co/docs/reference/integrations/zoom)

- [zscaler_zia](https://www.elastic.co/docs/reference/integrations/zscaler_zia)

- [zscaler_zpa](https://www.elastic.co/docs/reference/integrations/zscaler_zpa)

</details>

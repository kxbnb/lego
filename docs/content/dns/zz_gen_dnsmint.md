---
title: "DNSMint"
date: 2019-03-03T16:39:46+01:00
draft: false
slug: dnsmint
dnsprovider:
  since:    "v5.5.0"
  code:     "dnsmint"
  url:      "https://dnsmint.com"
---

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsmint/dnsmint.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->


Configuration for [DNSMint](https://dnsmint.com).


<!--more-->

- Code: `dnsmint`
- Since: v5.5.0


Here is an example bash command using the DNSMint provider:

```bash
DNSMINT_API_KEY="dnsm_xxxxxxxxxxxx_yyyyyyyy" \
lego run --dns dnsmint -d '*.q7k4m2.dnsmint-a3f9c1.dev' -d q7k4m2.dnsmint-a3f9c1.dev
```




## Credentials

| Environment Variable Name | Description |
|-----------------------|-------------|
| `DNSMINT_API_KEY` | API key with the dns01:write scope |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).


## Additional Configuration

| Environment Variable Name | Description |
|--------------------------------|-------------|
| `DNSMINT_API_URL` | API base URL (Default: https://dnsmint.com/api) |
| `DNSMINT_HTTP_TIMEOUT` | API request timeout in seconds (Default: 30) |
| `DNSMINT_POLLING_INTERVAL` | Time between DNS propagation check in seconds (Default: 2) |
| `DNSMINT_PROPAGATION_TIMEOUT` | Maximum waiting time for DNS propagation in seconds (Default: 60) |

The environment variable names can be suffixed by `_FILE` to reference a file instead of a value.
More information [here]({{% ref "dns#configuration-and-credentials" %}}).

## Description

DNSMint mints hostnames on domains it operates and serves from its own authoritative nameservers,
so the DNS-01 challenge is published through the DNSMint API rather than a zone you run.

The API key needs the `dns01:write` scope. It can be narrowed to one hostname or one domain when created.



## More information

- [API documentation](https://dnsmint.com/api-reference)

<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->
<!-- providers/dns/dnsmint/dnsmint.toml -->
<!-- THIS DOCUMENTATION IS AUTO-GENERATED. PLEASE DO NOT EDIT. -->

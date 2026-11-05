# What's next?

## Code cleanup

`hashy` is in a prototype state at the moment. It needs some refactoring to clean things up.

## Currently implemented

- Loading groups from one or more .zone files
- Building one hash ring for each group
- Responding to DNS queries with entries selected from hash rings

## Currently not implemented

- Loading groups from another data source, such as:
  - Another DNS server, e.g. Route53. This could involve zone transfers.
  - A database
- A bulk interface for doing reverse hashes (e.g. which devices should hash to a given talaria)
- Integration with existing services:
  - Petasos (existing DNS interface)
  - Scytale (existing DNS interface)
  - Talaria (unimplemented bulk interface)
- DNSSEC and security

## Open issues

- How to do zone transfers
- Should hashy be deployed via anycast, multicast, etc?
- How the bulk interface should be secured
- Rendezvous vs consistent hashing

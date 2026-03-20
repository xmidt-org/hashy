# Hashy Design

## Key Features

- Accepts DNS requests for devices
  - Host name is synthetic and contains the device name (e.g. MAC)
- Provides a RESTful interface that allows checking for updates
  - Hashy's [groups](#groups) can be updated dynamically
- Static configuration of groups
  - A group is a list of host names (i.e. talaria servers)

## Concepts

### DNS for devices

Hashy's DNS server understands hostnames with the following form:

```text
{deviceName}[-{ignored text}].{subdomain}
```

#### {deviceName}

The `deviceName` is the string that is hashed. It can only contain characters that are valid for a hostname. The meaning of the `deviceName` is opaque to Hashy. It can be a MAC address, UUID, or any arbitrary identifier. The list of servers returned via DNS is solely determined by the `deviceName`.

#### {ignored text}

Anything that comes after the `deviceName`, separated by a hyphen (`-`), is ignored. This allows devices to incorporate nonce values or other information for debugging and to ensure hostnames are not cached by intermediaries.

#### {subdomain}

The `subdomain` is a domain that Hashy will respond to. Hashy can respond to any of a list of subdomains, set via configuration.

Hashy will act as the SOA for configured subdomains. Any requests that don't match a configured subdomain results in an unknown response.

### Groups

Hashy organizes servers into `groups`. A *group* is simply *a list of servers with a unique name*. A group can be a datacenter, but it can also be any arbitrary list of servers. A server may belong to multiple groups. Groups are supplied to Hashy via configuration or (TODO) dynamically at runtime.

Hashy computes a hash of each group so that clients can determine if a group's members have changed.

## Flows

### CPE uses Hashy (instead of Petasos) to find a Talaria

```mermaid
sequenceDiagram
  participant CPE
  participant Hashy
  participant Talaria@{"type" : "collections"}

  CPE->>Hashy:DNS request
  Hashy->>Hashy:produces a hash, one per group
  Hashy->>CPE:DNS response containing a list of one IP or CNAME per group
  CPE->>CPE:chooses an IP or CNAME at random
  CPE->>Talaria:connects
```

### Talaria checks devices upon connection

```mermaid
sequenceDiagram
  participant CPE
  participant Talaria
  participant Hashy

  CPE->>Talaria:connects
  Talaria->>Talaria:Queues new connection for verification
  loop Submits queued connections to Hashy
    critical Verifies each CPE
      Talaria->>Hashy:queued device names
      Hashy->>Talaria:which device names hash to that talaria
      option Device is incorrectly connected
        Talaria->>CPE:disconnect with reason
    end
  end
```

### Talaria enforces device hashing

```mermaid
sequenceDiagram
  participant Talaria
  participant Hashy

  Talaria->>Hashy:(on startup) the host name for this talaria
  Hashy->>Talaria:groups for that host name, along with a hash for each group
  loop On a configurable interval, check Hashy for changes
    critical Check for updates
      Talaria->>Hashy:list of groups and hashes
      Hashy->>Talaria:updated groups and hashes
      option No changes
        Talaria->>Talaria:update TTLs (possibly)
      option Change detected
        Talaria->>Hashy:list of devices (possibly batched)
        Hashy->>Talaria:which devices must be relocated
        Talaria->>Talaria:remove devices which no longer belong
    end
  end

```

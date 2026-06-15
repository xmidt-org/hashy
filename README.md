# hashy

hashy is a distribute hashing-as-a-service server. It uses the results of hashing to direct traffic.

[![Build Status](https://github.com/xmidt-org/hashy/workflows/CI/badge.svg)](https://github.com/xmidt-org/hashy/actions)
[![codecov.io](http://codecov.io/github/xmidt-org/hashy/coverage.svg?branch=main)](http://codecov.io/github/xmidt-org/hashy?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/xmidt-org/hashy)](https://goreportcard.com/report/github.com/xmidt-org/hashy)
[![Apache V2 License](http://img.shields.io/badge/license-Apache%20V2-blue.svg)](https://github.com/xmidt-org/hashy/blob/main/LICENSE)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xmidt-org_PROJECT&metric=alert_status)](https://sonarcloud.io/dashboard?id=xmidt-org_PROJECT)
[![GitHub release](https://img.shields.io/github/release/xmidt-org/hashy.svg)](CHANGELOG.md)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/xmidt-org/hashy)](https://pkg.go.dev/github.com/xmidt-org/hashy)

## Summary

Summary should be a small paragraph explanation of what this project does.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Hash Protocol](#hash-protocol)
- [Install](#install)
- [Contributing](#contributing)

## Code of Conduct

This project and everyone participating in it are governed by the [XMiDT Code Of Conduct](https://xmidt.io/docs/community/code_of_conduct/). 
By participating, you agree to this Code.

## Hash Protocol

`hashy` provides a low-level protocol for hashing multiple objects at once and checking which objects hash to a subject. All integral values are big-endian. Binary strings are represented as a single length octet followed by the octets in the string.

Hash requests concern *subjects* and *objects*. `hashy` assigns *objects* to *subjects* via distributed hashing. A *subject* is often a host name, while an *object* is often a MAC address or similar identifier. *Subjects* and *objects* are represented as binary strings.

Hash requests may refer to *groups*. *Groups* are simply named sets of *subjects*. *Groups* are represented as binary strings.

### Header

```mermaid
---
title: "Hash header"
---
packet
+16: "Magic number (0xA9F4)"
+8: "Version"
+16: "Message ID"
+1: "RS"
+1: "ERR"
+6: "Message Type"
+32: "Message length"
```

#### Magic Number

All hash messages are prefixed with the 16-bit value **0xA9F4**.

#### Version

Version holds the 8-bit protocol version. The initial version is **1**.

#### Message ID

A client can set a 16-bit identifier to uniquely identify a message. `hashy` will place this same identifier into the response message.

#### RS bit

A single bit indicates whether this is a request or response.  **0** is used for request, **1** for response.

#### ERR bit

This bit is set if the message represents an error.

#### Message Type

Message types are 6-bit values that indicate the purpose and layout of the message.

| Value | Message | Description |
| --- | --- | --- |
| 000000 | Hash | Request contains one or more objects to hash. Response contains a map of objects to subjects. |
| 000001 | Reverse Hash | Request contains a subject and multiple objects. Response contains a list of objects that still hash to that subject and ones that do not. |

#### Message Length

A 32-bit length integer concludes the header and specifies how many message bytes follow.

### Error response

For error responses, the [message length](#message-length) will be set to **4** and the message will consist of a 4-octet error code.

### Hash

```mermaid
---
title: "Hash Request"
---
packet
+16: "Subject (variable length)"
+8: "Group count (0 indicates no filtering by group)"
+16: "Groups"
+8: "Object count"
+16: "Objects"
```

```mermaid
---
title: "Hash Response"
---
packet
+16: "Subject (variable length)"
+8: "Count of object/group entries"
+16: "Object (variable length)"
+16: "Group (variable length)"
```

### Reverse Hash

```mermaid
---
title: "Reverse Hash Request"
---
packet
+8: "Group count (0 indicates no filtering by group)"
+32: "Groups"
+32: "Subject (variable length)"
+32: "Object count"
+32: "Objects (variable length)"
```

```mermaid
---
title: "Reverse Hash Response"
---
packet
+32: "Subject (variable length, same as request)"
+32: "Keep count (objects that hash to the subject)"
+32: "Objects (variable length)"
+32: "Reject count (objects that DO NOT hash to the subject)"
+32: "Objects (variable length)"
```

## Install

Add details here.

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).

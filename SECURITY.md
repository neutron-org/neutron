# Neutron bug reporting

The Neutron core development team uses GitHub to manage feature requests and bugs. This is done via GitHub Issues.

## Standard priority bug

For a bug that is non-sensitive and/or operational in nature rather than a critical vulnerability, please add it as a GitHub issue.

If it is not triaged in a couple of days, feel free to tag `@swelf19`, `@albertandrejev`, `@oldremez`, `@oopcode`, `@pr0n00gler`, `@ratik`.

## Critical bug or security issue

**NOTE: this is a temporary policy. We are currently setting up an Immunefi bug bounty program. This document will be updated soon.**

If you're here because you're trying to figure out how to notify us of a security issue, get in touch via email [security@neutron.org](mailto:security@neutron.org) or alert us on Telegram:

- @andrei_zavgorodnii (Andrei Zavgorodnii)
- @Spaydh (Avril Dutheil)
- @kayvflu (Kai Tiurin)
- @pr0n00gler (Mike Mozhaev)
- @sw31f (Sergei Golyshkin)
- @nnnooo1111 (Sergei Ratiashvili)
- @Albert_OpenTech (Albert Andreev)

Please avoid opening public issues on GitHub that contain information about a potential security vulnerability as this makes it difficult to reduce the impact and harm of valid security issues.

### Coordinated Vulnerability Disclosure Policy

We ask security researchers to keep vulnerabilities and communications around vulnerability submissions private and confidential until a patch is developed. In addition to this, we ask that you:

- Allow us a reasonable amount of time to correct or address security vulnerabilities.
- Avoid exploiting any vulnerabilities that you discover.
- Demonstrate good faith by not disrupting or degrading Neutron’s network, data, or services.

### Vulnerability Disclosure Process

Neutron uses the following disclosure process:

- Once a security report is received, the Neutron core development team works to verify the issue.
- Patches are prepared for eligible releases in private repositories.
- We notify the community that a security release is coming, to give users time to prepare their systems for the update. Notifications can include Discord messages, tweets, and emails to partners and validators.
- 24 hours following this notification, the fixes are applied publicly and new releases are issued.
- Once releases are available for Neutron, we notify the community, again, through the same channels as above. We also publish a Security Advisory on Github and publish the CVE, as long as neither the Security Advisory nor the CVE include any information on how to exploit these vulnerabilities beyond what information is already available in the patch itself.
- Once the community is notified, we will pay out any relevant bug bounties to submitters.
- One week after the releases go out, we will publish a post with further details on the vulnerability as well as our response to it.

This process can take some time. Every effort will be made to handle the bug in as timely a manner as possible. However, it's important that we follow the process described above to ensure that disclosures are handled consistently and to keep Neutron and the projects running on it secure.

*P.S. Many thanks to the Juno community for an awesome template for this document!*
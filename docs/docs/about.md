# About kod

The Kubernetes Offline Deployment (kod) tool allows you to package a helm chart or helmfile and their dependencies
and publish them to a Kubernetes cluster in an offline, air-gapped, environment.

This tool stays as close to `helm` and `helmfile` as possible, and tries not to introduce any side effects in how
it operates. It does not hamper the deployer at all by restricting namespaces, container registries, or any
other aspect of a deployment. It does not install any additional webhooks or CRD into the target Kubernetes cluster.
All commands executed are outputted to the console, so you can choose to execute them yourself in future.

kod will prefer the OCI container images and artifact format, but does support the legacy dockerv2 format but will
use whatever format it finds in the source URL.
`reglclient` is used for OCI registry interactions, and `skopeo` is used for dockerv2 registry interactions.

## Motivation

Having worked on private, air-gapped installations on numerous occassions the last 7 years and finding existing
tooling too complex, I decided to create a simple tool that does one thing and one thing only - make it easy
to move helm charts and containers from a public container registry environment into a private container
registry environment, and deploy them using standard helm (and helmfile) tooling. This follows the Unix methodology
that each tool should do one thing well, and only that thing.

Existing approaches start simple, but wind up being very complicated. They force you to change your CI/CD environments
and create complex configuration files, changing the way you handle variables and values files, and placing themselves
between you and executing helm commands. They even modify the way you use a live Kubernetes system, forcing you to
run additional workloads in each and every cluster, introducing their own CRDs and webhooks, and making the security
boundary (and thus blast radius) of your target environments that much bigger.

kod doesn't force you to do this. kod will enable you to package and move containers and helm charts to an
offline, air-gapped environment, but do so in a way that is fundamentally helm native. The acid test of this
being your ability to run helm commands on their own once kod has been used to move your charts and containers
to an offline container and OCI artifact registry. kod always outputs the CLI commands it uses in its own logging,
to build your knowledge in the needed commands, for `skopeo` and `helm` command invocations.

## About the name

Being from Scunthorpe, I couldn't pass up the opportunity to continue a long-running inter-town rivalry joke with
our fishing cousins from nearby Grimsby! We call those fine folk from there 'cod heads'... even though they fish for Haddock
and not Cod... We do this to see their angry reaction to us apparently confusing the type of fish they catch.
It's all good fun.

![Kod Logo](images/logo.png)

This is why the logo for 'kod' is, in fact, a Haddock. 8o) It's a laugh, innit? Up the Iron! Oh and go order some
Haddock from the fish and chip shop, not Cod, as Grimsby Haddock is much better. 
[Haddock is also more sustainable than Cod](https://www.mcsuk.org/goodfishguide/species/haddock/) for British sourced fish.
(Although I've not eaten fish in 5 years since watching [Seaspiracy](https://www.seaspiracy.org/) - Content Warning on that film).

## Technical approach

A few key principles are at play in kod:-

- Don't modify the target k8s environment
- Handle any complexity as much as we can in the kod command itself
- Allow overrides in a helm-friendly way where we need to customise behaviour to a package (E.g. when they use weird value names for container registry, repository, tag and digest)
- Wrap other tools that do jobs well rather than create our own (E.g. skopeo, helm, helmfile, and the zot container registry)
- Target the OCI container image format and artifact format and protocols as much as possible (COMING SOON IN v0.6.0)
- Ensure packages are self-describing, requiring no external information to 'just work'
- Don't assume our container image sources are always public - allows a helm values override on packaging to target a local container registry (E.g. to fetch pre-scanned, authorised container images)

Some principles are also at play to help users learn the system, rather than hide complexity in the tool:-

- Always show the output of skopeo, helm, and other CLI invocations so they can be performed without kod if needs be
- Allow introspection of the temporary folder and archive format, including via the `unpack` command - allows scanning before deployment
- Minimise the amount of file operations by not always overwriting temporary files if they exist - allows for resuming partial broken package/deploy commands
    - Also, always allow a --cleanup command option to be used, to clear away these files

## Packages known to work with kod

Any helm chart that uses common names for container image values in the values file should work. Note that this
defaults to image.repository and image.tag as a minimum, but could also split the registry out into image.registry,
and may also specify a specific container image SHA256 hash in an image.digest property.

See the [Validated Charts page](./reference/validated.md) for a comprehensive list.

Other automated mechanisms for finding container image references will be added over time. You will
be able to override this in future for troublesome packages using a kod-hints.yaml file.

The kod project will try and ensure that at least the top ten monthly viewed charts on artifacts hub always
work OOTB. See [https://artifacthub.io/stats](https://artifacthub.io/stats) 
(top right) for that list. We will also provide a mechanism and
how-to for getting any helm chart working with kod using the hints file.

## Tools we use

Rather than re-invent the wheel, we use and support a number of other opensource tools:-

- [`skopeo`](https://github.com/containers/skopeo) for downloading and uploading container images via the docker v2 protocol
- [`helm`](https://helm.sh/) for templating, packaging, and deploying the helm charts
- [`helmfile`](https://helmfile.readthedocs.io/en/latest/) (Soon!) for sequencing the installations of multiple helm charts across namespaces
- [`zot`](https://zotregistry.dev/) container and artifact registry as our target OCI registry for testing and validating kod. Other container registries are supported though

Other tools are available... We just don't use them with kod. (E.g. `docker save`)

## FAQs

### Why use helm?

Helm is the de-facto standard for packaging and versioning a Kubernetes application. Other mechanisms such as git
repos of Kustomize scripts are just painful.

### Why use helmfile?

The main weakness of helm itself is that each deployment can only target a single namespace. Chart dependencies
are aimed at real dependencies, not for installing chains of charts. helmfile allows you to order the
installation of helm charts across multiple namespaces, and stays native to helm throughout, making it a
good partner to kod.

helmfile also supports installing each chart multiple times, with different settings, which can
be very useful. E.g. for an Istio installation which deploys the `gateway` chart multiple times in a single
mesh across multiple namespaces.

### Why not use &lt;MY-FAVOURITE-TOOL&gt; instead of kod?

I've tried your favourite tool. It was painful. It changed me, and not in a good way (and my entire DevOps pipeline
along with it). kod is simpler and tries to 'just work' OOTB with a couple of commands and without introducing any
additional technological complexity.

### Why don't you support Windows?

We don't support Windows purely because the 
[`skopeo` tool does not support Windows](https://github.com/containers/skopeo/issues/715#issuecomment-3649231030)
(Although this is in the works at the moment... Check the link). Once that does become available, we will support it.
We do build kod on Windows and publish that binary, just in case one day it does become available. On that day this
`kod.exe` programme will work straight away... Assuming Windows' skopeo isn't broken of course...

### Do you provide FIPS binaries?

Bilberry Software Ltd can provide FIPS binaries for a small charge to those that need them. As we invoke
other tools, and those tools are the ones that use openssl and other cryptographic suites, we'd supply FIPS
builds of those tools and just invoke them with kod.

### Why write kod in GoLang?

Go is awesome, and is well understood in the Kubernetes community. It also makes for lightweight standalone
binaries which can be easily installed with the *nix `install` command. I quite often package up all the K8s
related CLI tools into a single archive along with an installation script that invokes install for each binary.
This approach makes for a very simple way to provision jump hosts and dev/test/qa Ubuntu VMs.

We also may want to allow people to plug in extensions in the future by creating a `kod-COMMAND` exe separately.
This is easy to do and quite common in GoLang CLI applications. (E.g. kubectl extensions)

### How do you do testing?

We plan to use the [Concourse](https://concourse-ci.org/) CD tool for all testing, and to rebuild and retest when an upstream dependency
changes. This lets us be completely hands-off for minor patch builds, whilst supporting and testing the latest
versions of our tooling. We'll provide a link to our public pipelines once crafted. We will also use GitHub actions,
but this is purely to carry out GoLang code checks and run unit tests for the CLI tool itself, not perform
release build creation or carry out complex integration testing. Concourse is better suited to that.

This is possible because Concourse is a declarative, dependency driven CD tool and builds when any upstream
dependency changes. It doesn't rely on a code update from ourselves to instigate any of this version
compatibility testing, instead watching for any upstream change, not just a change to our source code.

Concourse is a 'generic thing runner' and can run on any platform, natively or in Kubernetes, providing a very
rich build and test environment. We have Concourse runners on Windows, Mac, and Linux Ubuntu automating the creation
and testing of our software against Kubernetes clusters and Ubuntu VMs.

### What's your background and why the choice of technologies?

Our main author, [Adam Fowler](https://github.com/adamfowleruk/), has worked for years in regulated industries,
including healthcare, financial services, insurance, and national security/defence.
Since 2018 he's mainly worked with Kubernetes and other containerisation technologies, initially working at
Pivotal/VMware, advising organisations on their Kubernetes, DevOps platform automation, and cloud native software
development practices and architecture.

Other claims to fame/notoreity include writing NoSQL for Dummies in 2015 and being the lead architect of the
England & Wales COVID-19 Digital Contact Tracing app between March and June 2020 before going away and starting the
[Herald Proximity](https://heraldprox.io/) opensource Bluetooth library used in the Australian and Alberta, Canada
COVID-19 Digital Contact Tracing apps. This is still being actively maintained and is being used as part of Adam's PhD in Clinical
Medicine at Oxford University.

The technology choice is purely a practical one, using the de-facto standard technologies for application
definition and packaging on Kubernetes, and providing a simple tool that slots into them without changing
how you use them. My general approach is to meet the customers/end-users where they are, rather than trying
to force changes in behaviour that cost time in training and can lead to greater maintenance and security risks
down the road, if they are poorly understood or used.

I used Concourse when at Pivotal and fell in love with it. It allows very small teams to maintain
large complex software products in a mostly hands-off way. The same with VMware Kubernetes Service
(formerly Tanzu Kubernetes). I've also been using Microk8s the last 4 years on Ubuntu Server, and love the
strict confinement mode of it, and its simplicity. Zot too works really well as an Ubuntu Service or as a K8s
deployment in a shared services cluster. It's super lightweight and easy to use, and fast, and has improved
tremendously over the last 2 years.

At Bilberry Software Ltd we can provide support, guidance, consulting, and FIPS builds of all of the above,
configured to UK Government/NCSC and CIS Benchmark standards.

# Kod architecture

This file describes guiding principles and the architectural decisions that fall out of them. This gives some
indication of the future direction of travel, and the kinds of features we seek to add and will never have.

kod packages helm charts and helmfiles and the container images they reference for offline deployment, 
using other tools to get the job done.
In particular, we wrap two ways of describing a K8s application, both orientated around Helm:-

- [Helm](https://helm.sh/) - the de-facto standard way to describe a K8s application
- COMING SOON: [helmfile](https://helmfile.readthedocs.io/en/latest/) - A way to wrap multiple Helm charts for deployment to different namespaces, with settings overrides. This wraps helm too.

## Unix philosophy of one tool for one job

We want kod to be the best and easiest way to package Helm Chart and helmfiles and get them installed on
an air-gapped or edge environment. All functionality should make this the aim, and we do nothing beyond this.

In particular, this means that we intend to:-

- Not have any 'runtime' components installed in a Kubernetes environment to support deployment or upgrades
- Make your CI/CD runs simpler, not change them to suit our tool
  - I.e. if you today manually fetch helm charts and the required containers, you can simplify that by using kod, rather than change your whole packaging philosophy away from pure Helm and helmfiles
  - We may provide a container image distribution of zot with tools to make CI steps easier, but I doubt we'll add specific support to specific CI/CD tools
- We will only wrap other tools, not try to borg them into our binary, with all the version and security issues that brings with it
- If something makes consuming kod difficult, then we may provide a command to ease that
  - Like installing a container registry in an airgap, by providing a package-zot and deploy-zot command
  - Like provide a way to package kod and the tools it wraps, by providing a package-tools and deploy-tools command
  - But ONLY when they are direct pre-requisites to getting up and running
    - We won't support setting up every container registry, just an example one: zot

## We build around Helm

Helm has become the de-facto standard for describing, packaging, and distributing a Kubernetes application.
We intend to fully get behind this package standard, warts and all, to encourage the community to coalesce around
a single application description format.

We do also support helmfile. Isn't that a contradiction? No it isn't because helmfile itself too just wraps
Helm rather than trying to replace how it works. Kod being able to wrap an individual Helm chart and a group of
Helm charts via a single Helmfile solves most of the limitations of using Helm on its own.

Note also that this implies we don't support the following scenarios:-

- Packaging multiple Helm charts in the same kod package
  - helmfile already handles the myriad of multi-chart configuration, so why would kod re-implement this?
  - Just create a helmfile and package this for offline use with kod
- Packaging multiple helmfiles in the same kod package 
  - You should generate a single kodpackage per helmfile for simplicity
- Having an Uber-Package format that tries to wrap multiple kod packages
  - Other tools that do this tend to tie themselves in knots, and have poor install/uninstall repeatability
  - They tend to also introduce massive complexity in mapping variables from those formats, to lower helm charts, and across them
    - This makes them restrictive in the amount of settings overrides you can have in lower level packages, removing the original Helm author's flexibility and options from the end user
- We don't support directly using Kustomize or raw kubectl YAML files
  - Seriously, helm is a much better way of handling this, and has the benefit of ensuring your app is *versioned* properly, allowing repeatable builds

## We don't get in the way of Helm

You should be able to package any Helm chart of helmfile and still have access to all command options of those tools.

kod will not restrict you from that flexibility. kod will provide options like `unpack` which unpacks a kod package so you
can execute your helm commands on them directly, and options like `deploy --list-commands` to output what deploy helm command lines it would execute,
allowing you to take them and add in your own flags and additional options, if necessary.

As an example, kod provides pass-through support for these flags on deploy:-

- `-n MYNAMESPACE` to specify the installation K8s namespace
- `-d MYDEPLOYMENTNAME` to specify the helm deployment name
- `-f VALUES1.yaml -f VALUES2.yaml ...` to specify additional override helm values files
  - Note that kod prepends its own to this list, to specify the new private registry location of the copied container images it has deployed for you
- COMING SOON: `--wait` to instruct Helm and helmfile to wait for the installation to complete and become healthy before succeeding

If you need more options than this, you can do the following:-

```shell
kod unpack -c kod-mychart-myversion.tgz
kod deploy --no-install --list-commands -c /tmp/kod-mychart-myversion/ -r https://myregistry:8080/ -n mynamespace
```

The above two commands will:-

- Unpack the chart kod package to a temporary folder (This command is actually optional, as deploy will unpack it anyway)
- Install the container images in the local airgapped container registry
- Generate the /tmp/kod-mychart-myversion/charts/mychart-myversion-values.yaml file(s) that you should add to the helm commands
  - These container the pointer to the local container registry container image URLs and tags
- Output the full list of helm upgrade --install commands with pointers to our own values files and options

You can then simply modify these commands to add any additional options you wish, then execute them yourself. This
gives you access to all the helm command options without kod hiding or restricting what you can accomplish.

The above approach is also a great way to plug in security inspections and automated tooling to inspect the packaged
contents without having to instruct those tools in how to handle the kod CLI options or the kodpkg archive format.

## Tool architecture

We have chosen GoLang because it's awesome and interacts well with the rest of the K8s ecosystem which is also GoLang based.

We try to follow conventions in other tools such as kubectl and helm where they make sense. In particular, if we need to support
a helm flag (like `-f` to point to one or more custom values files) then we adopt the same command flag names and abbreviations.

## OCI and Zot

The OCI packaging format is moving beyond wrapping container images, and to describing generic deployable artifacts.
The Zot registry is building built from the ground up to support the OCI format. We've been really happy
watching how Zot has matured over the past 3 years and have used it in anger on a number of Edge and airgapped
installation projects, achieving great results.

kod aims to follow OCI standards wherever possible, and integrate their OCI features not just for container image
distribution, but we also allow the fetching helm charts via OCI URLs in the package command, and will in future
modify the deploy command to push the helm charts themselves and kodpkg or kodinfo files as artifacts too.
That will effectively mean after the first deploy or populate command of a kodpackage, you should be able to
execute kod against the OCI registry with no local files stored on disc at all.

### If you need an online container registry are you really an offline deployment tool?

Yes absolutely. The CIS benchmark for Kubernetes requires that all container images are validated against
a container registry on start. This prevents modification on disc or injection of any malware.
This means you always have a container registry accessible from a K8s cluster.
Whether this is on-cluster or off-cluster doesn't matter - you always have one. 

Other offline deployment tools do deploy their own 'embedded' container registry, which is normally only for storing container 
images and not helm charts as artifacts, usually using the `docker-registry` container image which does not
support the OCI specification fully.

Rather than try and boil the ocean by deploying our own, we instead allow you to tell us where your local one
is, and we populate and use that for deployments. It could be running anywhere - on-cluster or off-cluster.
We're not container registry developers though, and don't pretend we know how you should run one.

If this registry happens to support OCI artifacts, we will even install the helm charts and kodinfo packages
themselves as artifacts, allowing kod to then run fully against the registry with no local file storage.
This provides additional security in that artifacts can be signed, preventing malicious modification on disc.

We do plan, however, to allow you to package and deploy the zot container registry via the `package-zot`
and `deploy-zot` commands. This is purely as a convenience and example. How you then operate that container
registry is down to you. We deploy it off-cluster on a single Ubuntu VM as a systemd service. It may well
be that this Ubuntu VM is also running Kubernetes, but running the registry outside of K8s has a number of 
benefits, not least of which is not bypassing the aforementioned CIS benchmark checks on the container 
registry's own container image.

## Why should I care about the kod approach?

A few reasons:-

- A smaller binary is a smaller surface area for security issues
  - We don't wrap and hide the other tools' binaries - you can inspect them and their sha256 hash
  - You're not limited by our version or flag support of these tools either, allowing you to adopt the latest and most secure versions immediately
- You don't have to rework security scans, build pipelines, how you handle helm config, just to use kod
  - This reduces training time
  - It reduces the likelihood of mistakes through learning yet another tool's different way of doing the same thing
  - This increases security through avoiding mistakes or misinterpretations
- You don't have to break your K8s cluster or security posture just to support our tool
  - No need to re-write every Pod description via a webhook to use an internal container registry, breaking hybrid helm support on that cluster
  - No need to introduce permanent additional components into your cluster that you'd have to patch and keep updated, and which eat resources
  - We don't introduce any services that are outside your cluster's mesh, that don't support mTLS, and don't implement best practice around container non-root, unprivileged pod policies

## Isn't supporting container image overrides in helm properties waaaaay harder than overriding K8s pod image pointers?

Yeah, it is a fair bit harder, but I wouldn't say way harder. Certainly MUCH less difficult than implementing a bunch of k8s
apps and webhooks just to override the `image: ` property on pods at deployment time, or in the alternative approach of
scripting the use of a set of complex repackaging CLI tools and introducing yet another registry for apps.

We also have heuristics to find container image references in helm charts' default values.yaml files, and where that
doesn't work, you can provide a hint file per chart to tell kod where these properties live, and it'll handle the values file
and container copying as normal - there's no big cliff edge where there are helm packages we can just not support.

## What about the situation where I need to add a tweak to what is deployed, or perform some earlier action in K8s?

There are times you find yourself deploy 1-2 quick YAML files to workaround helm chart limitations. These tend to
grow to 10 YAML files.
Most of the time you should separate out platform deployment concerns - configuring the K8s environment ready to 
receive an application by installing tooling such as cert-manager, operators et al - and 
app installation concerns - how many and how large a database I want, and what application(s) will consume them.

This is easier said than done though. It's not really something kod can fix, as its down to Helm chart and
Operator authors, but below are a couple of patterns that will work and that kod could help you package and use.

### Platform layer customisation

A Platform layer is a layer of K8s functionality added to an empty vanilla K8s cluster that provides
required functionality pre-requisites to an application installation. Examples include cert-manager for cert generation, 
Istio for ingress, or a variety of Data operators to allow for 'easy' instantiation of databases.

You may need to 'instantiate' certain objects or K8s CRDs after a platform layer is installed, but before
an application is deployed. This is where you may be tempted to just use a couple of small YAML files.

Most of these issues stem from platform tooling using Helm as packaging themselves. E.g. You can install a
cert-manager package, but then you may need to create `Issuer`s or `ClusterIssuer`s. You may perceive us at kod saying
you need to create a wrapper Helm chart as complex overhead. In which case I'd say you should do the following:-

- Define your platform layer using Helmfile, E.g. installing a set of components including Helmfile, Istio etc. via Helm
- Define a single MyPlatformLayerConfiguration Helm chart, which configures ALL of these components. E.g. you may create multiple cert-mamnager `Issuer`s and refer to them in multiple Istio `Gateway` definitions
- Add this configuration helm file as the last ordered helm chart in your platform layer's helmfile. (or, more likely, as the first helm chart in the next layer in the stack)

This means you produce only one 'runtime configuration' package per platform layer, which is reasonable as most of the
config you generate will relate to one another, so a single Helm chart makes more sense than one config chart per helm chart,
and will be much less work for you.

You then simply create multiple layers and install them in order. Or indeed with the above, create a single
helmfile that just ensures the platform components and config helm charts are just installed in the correct order
using `--wait` option in helmfile config where appropriate.

You can then simply use kod to wrap the helmfile, preserving all the configuration and command options that
you want, but with the knowledge it'll all deploy to an air-gapped environment fine, and you don't have to do
or maintain any weird kod-specific workarounds - you're just in Helm land the whole time.

This way you end up with only 1-3 helmfiles and thus 1-3 helm packages, for one 'required' and, say, 2 'optional'
platform layers.

### Data layer Operator instantiation helpers

Another area I've seen this is in the Operators Pattern. Operators are meant to simplify the instantiation
of data layers on Kubernetes. Unfortunately, most of the time, they simply move complexity into the Operator
layer and out of the helm layer. Have you ever tried using a Nifi, Kafka, Elasticsearch, or MongoDB operator?
It's HELL. That YAML basically requires a K8s app developer to be an expert in how those data layers are architected
in order to achieve any security, performance, or high availability.

(Except the [RabbitMQ Operators](https://www.rabbitmq.com/kubernetes/operator/operator-overview. They are awesome.)

What most people end up doing is creating another Helm chart - one per operator - to spit out the Operator CRD with working configuration.
This of course defeats the entire purpose of the Operator pattern, and is in my opinion a SEVERE weakness in the K8s
landscape. It's not the operator pattern's fault, but rather the operator writer forgetting who the primary user is - 
an application developer that is NOT an expert in the data layer they wish to deploy.
Some projects get this right (E.g. RabbitMQ), but most data layer operators are just painful to work with.

At this point you may of course just as well use the Helm chart for the data layer you have the operator for,
then use a set of well known values files to deploy working configuration as overlays. It would be better if
Operators were all  
[simple wrappers around Helm charts](https://sdk.operatorframework.io/docs/building-operators/helm/tutorial/), 
and the operator CRDs provided just a high level consumer API rather
than replicating a Helm values file low-level API and making the complexity arguably worse, exposing it all
to a K8s app developer. This would make the Operator pattern more useful.

COMING SOON: the `package` command in kod will support including your own values files in the package alongside
the ones we generate that override the image locations. This allows you to effectively specify, for example,
a default working MongoDB configuration via its Helm chart, leaving deployers with simpler high-level overrides 
in Helm such as the number of MongoDB instances they wish to run.

## In summary

All of the above solutions are possible because we use each tool for its appropriate job, and don't introduce
complexity through kod-specific solutions, workarounds, or tweaks to Helm or helmfile.

Does this mean you have some work to do? Maybe, but its limited to just creating a small Helm chart
in some very specific, one off, circumstances. Once they're created you'll probably never have to touch them
again. You'll also not have to learn any kod-specific tweaks or workarounds to what are, fundamentally,
solutions to be achieved with Helm and helmfile.